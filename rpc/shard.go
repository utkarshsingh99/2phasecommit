package rpcmanager

func (r *RPCManager) UpdateShardMapping(clusterID string, serverID string) {
	r.ShardLeaderMapping[clusterID] = serverID
}
