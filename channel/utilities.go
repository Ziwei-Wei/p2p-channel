package channel

const (
	unixWeek   int64 = 604800
	unixDay    int64 = 86400
	unixHour   int64 = 3600
	unixMinute int64 = 60
)

func (channel *ChatChannel) getPeerLatestMsgID(peerName string) (int, error) {
	var info peerMessageInfo
	err := channel.db.
		From(channel.channelName).
		One("PeerName", peerName, &info)
	if err != nil {
		return -1, err
	}
	return info.RecorderInfo[peerName].LatestMsgID, nil
}

func (channel *ChatChannel) updatePeerLatestMsgID(peerName string, newMsgID int, lastSyncTime int64) error {
	err := channel.db.
		From(channel.channelName).
		UpdateField(&peerMessageInfo{PeerName: peerName}, "RecorderInfo", newMsgID)

	if err != nil {
		return err
	}
	return nil
}

func (channel *ChatChannel) getAllLatestMsgIDs() ([]peerMessageInfo, error) {
	var info []peerMessageInfo
	err := channel.db.
		From(channel.channelName).
		All(&info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
