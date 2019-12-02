package space

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/asdine/storm/v3"
	"github.com/gorilla/websocket"

	channel "github.com/Ziwei-Wei/cyber-rhizome-host/channel"
	"github.com/Ziwei-Wei/cyber-rhizome-host/keygen"
	"github.com/Ziwei-Wei/cyber-rhizome-host/network"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Space > channels > topics
// Host, pubsub => Space(Team), host with a certain private key and name defines the id in the group
// subscriptions  => Topics, a subscription is a topic in the space
// host will connect all other know host when started then subscribe and listen for topics.
// there should always be a public topic with everybody in it to exchange and sync space info.
// Space should be a private network not allowing people without a space key to enter.

/*
0  init data structures to record data, read from files to set initial value
1  create host
2  create pubsub
3  subscribe topics and set validator
4  connect to known hosts
5  start listening
*/

// will go into caching mode when no peer in the channel is online

/* RhizomeSpace */

// RhizomeSpace main cordinator
type RhizomeSpace struct {
	ready    bool
	port     uint16
	space    string
	user     string
	host     host.Host
	pubsub   *pubsub.PubSub
	channels map[string]channel.Channel
	db       *storm.DB
	gui      *websocket.Conn
	mutex    sync.Mutex
}

// NewRhizomeSpace init the RhizomeSpace
func NewRhizomeSpace() *RhizomeSpace {
	return &RhizomeSpace{}
}

// Start the service
func (space *RhizomeSpace) Start(_port uint16, dbPath string) error {
	// connect to local data base
	db, err := storm.Open(dbPath)
	if err != nil {
		return err
	}
	space.db = db

	// start gui websocket
	http.HandleFunc("/space", func(w http.ResponseWriter, r *http.Request) {
		space.listenToGUI(w, r)
	})
	space.ready = true
	return nil
}

// LoginSpace login using private key, name, and corresponding network protector
func (space *RhizomeSpace) LoginSpace(spaceName string, userName string, protString string) error {
	protKey, err := keygen.StringToProt(protString)
	if err != nil {
		log.Println(err)
		return err
	}
	privKey, err := keygen.CreatePrivKey(userName, protString)
	if err != nil {
		log.Println(err)
		return err
	}

	// read localDB check if user exist
	var currSpace spaceInfo
	var existUser bool
	err = space.db.One("spaceName", spaceName, &currSpace)
	if err != nil {
		return err
	}
	for member := range currSpace.memberToAddrs {
		if member == space.user {
			existUser = true
			break
		}
	}

	/* if exist */
	if existUser == true {
		// 1  create host
		host, err := network.CreateHost(privKey, protKey, space.port)
		if err != nil {
			return err
		}

		// 2  create pubsub
		pubsub, err := network.CreatePubSub(host)
		if err != nil {
			return err
		}

		// 3 open channels
		for cName, cType := range currSpace.channelToType {
			curr, err := channel.OpenChannel(cName, space.user, cType, space.db, host, pubsub)
			if err != nil {
				log.Println(err)
				continue
			}
			space.channels[cName] = curr
		}

		// 4  finish
		space.host = host
		space.pubsub = pubsub

		return nil
	}
	/* if not exist */ // return error
	return errors.New("user doesn't exist")
}

//!!!!!!!!!!!!
// CreateSpace will create a space
func (space *RhizomeSpace) CreateSpace(spaceName string, privKey crypto.PrivKey) (string, error) {
	if space.ready != true {
		return "", errors.New("start rhizome spacefirst")
	}

	// create protector
	protString, err := keygen.CreateProtString()
	if err != nil {
		return "", err
	}
	prot, err := keygen.StringToProt(protString)
	if err != nil {
		return "", err
	}

	// create host to get curr multiaddress
	host, err := createHost(privKey, prot, space.port)
	if err != nil {
		return "", err
	}
	multiAddrs := getFullAddresses(*host)
	var memberToAddrs map[string][]string
	memberToAddrs[userName] = multiAddrs

	// save to localDB
	newSpace := spaceInfo{
		spaceName:     spaceName,
		memberToAddrs: memberToAddrs,
	}
	err = space.localDB.Save(&newSpace)
	if err != nil {
		return "", err
	}

	// save to orbitDB
	newReq := orbitRequest{
		ReqType:     "create",
		ContentType: "space",
		Content:     newSpace,
	}
	err = space.orbitDB.WriteJSON(&newReq)
	if err != nil {
		return "", err
	}

	return stringProt, nil
}

//!!!!!!!!!!!!
// JoinSpace will join an existing space with a referer
func (space *RhizomeSpace) JoinSpace(spaceName string, privKey crypto.PrivKey) {}

// End end the service
func (space *RhizomeSpace) End() error {
	space.db.Close()
	return nil
}

// Send will send message to given channel in current space
func (space *RhizomeSpace) Send(channel string, message string) error {
	return nil
}

// Logout logout current space
func (space *RhizomeSpace) Logout() {
}

// GetChannels get current peers in the channel
func (space *RhizomeSpace) GetChannels() []string {
	keys := make([]string, 0, len(space.channels))
	for key := range space.channels {
		keys = append(keys, key)
	}
	return keys
}
