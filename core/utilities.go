package core


// db level
type userInfo struct {
	userName      string `storm:"id"`
	spaceSet      map[string]struct{}
	friendToAddrs map[string][]string
	friendToPID   map[string]string
}

type spaceInfo struct {
	spaceName     string `storm:"id"`
	channelToType map[string]string
	memberToPID   map[string]string
	memberToAddrs map[string][]string
}
