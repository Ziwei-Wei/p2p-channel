package space

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/beevik/ntp"
	"github.com/gorilla/websocket"

	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	pnet "github.com/libp2p/go-libp2p-pnet"
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

/* RhizomeCore */

// RhizomeCore main cordinator
type RhizomeCore struct {
	ready    bool
	port     uint16
	space    string
	user     string
	host     *host.Host
	pubsub   *pubsub.PubSub
	channels map[string]*Channel
	db       *storm.DB
	gui      *websocket.Conn
	mutex    sync.Mutex
}

// NewRhizomeCore init the RhizomeCore
func NewRhizomeCore() *RhizomeCore {
	return &RhizomeCore{}
}

// Start the service
func (core *RhizomeCore) Start(_port uint16) error {
	// connect to local data base
	db, err := storm.Open("db/rhizome.db")
	if err != nil {
		return err
	}
	core.db = db

	// start gui websocket
	http.HandleFunc("/core", func(w http.ResponseWriter, r *http.Request) {
		core.listenToGUI(w, r)
	})
	core.ready = true
	return nil
}

// LoginUser check local db for saved user
func (core *RhizomeCore) LoginUser(userName string) {}

// LoginSpace login using private key, name, and corresponding network protector
func (core *RhizomeCore) LoginSpace(spaceName string, privKey crypto.PrivKey, protKey string) error {
	binaryProt := protStringToBinary(protKey)
	prot, err := pnet.NewV1ProtectorFromBytes(&binaryProt)
	if err != nil {
		return err
	}

	// read localDB check if user exist
	var currSpace spaceInfo
	var existUser bool
	err = core.db.One("spaceName", spaceName, &currSpace)
	if err != nil {
		return err
	}
	for member := range currSpace.memberToAddrs {
		if member == core.user {
			existUser = true
			break
		}
	}

	/* if exist */
	if existUser == true {
		// 1  create host
		host, err := createHost(privKey, prot, core.port)
		if err != nil {
			return err
		}

		// 2  create pubsub
		pubsub, err := createPubSub(*host)
		if err != nil {
			return err
		}

		// 3 open channels
		for channel, channelType := range currSpace.channelToType {
			currChannel, err := openChannel(channel, core.user, channelType, host, pubsub, core.db, core.gui, &core.mutex)
			if err != nil {
				log.Println(err)
				continue
			}
			core.channels[channel] = currChannel
		}

		// 4  connect to known hosts in space
		connectWithPeers(host, currSpace.memberToAddrs)

		// 5  finish
		core.host = host
		core.pubsub = pubsub

		return nil
	}
	/* if not exist */ // return error
	return errors.New("user doesn't exist")
}

// CreateSpace create a space with a name
func (core *RhizomeCore) CreateSpace(spaceName string, privKey crypto.PrivKey) (string, error) {
	if core.ready != true {
		return "", errors.New("start rhizome core first")
	}

	// create protector
	stringProt, err := createStringProt()
	if err != nil {
		return "", err
	}
	binaryProt := protStringToBinary(stringProt)
	prot, err := pnet.NewV1ProtectorFromBytes(&binaryProt)

	// create host to get curr multiaddress
	host, err := createHost(privKey, prot, core.port)
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
	err = core.localDB.Save(&newSpace)
	if err != nil {
		return "", err
	}

	// save to orbitDB
	newReq := orbitRequest{
		ReqType:     "create",
		ContentType: "space",
		Content:     newSpace,
	}
	err = core.orbitDB.WriteJSON(&newReq)
	if err != nil {
		return "", err
	}

	return stringProt, nil
}

func joinChannels(ps *pubsub.PubSub, channels []string) (map[string]*pubsub.Subscription, error) {
	var subscriptions map[string]*pubsub.Subscription
	for _, channel := range channels {
		sub, err := joinChannel(ps, channel)
		if err != nil {
			return nil, err
		}
		subscriptions[channel] = sub
	}
	return subscriptions, nil
}

// Logout logout current space
func (core *RhizomeCore) Logout() error {
	return nil
}

// End end the service
func (core *RhizomeCore) End() error {
	core.localDB.Close()
	return nil
}

func (core *RhizomeCore) syncTime() error {
	ntpTime, err := ntp.Time("time.apple.com")
	if err != nil {
		return err
	}
	currTime := time.Now()
	core.timeOffset = ntpTime.UnixNano() - currTime.UnixNano()
	return nil
}

// SendMessage will send message to given channel in current space
func (core *RhizomeCore) SendMessage(message string, channel string) error {
	return nil
}
