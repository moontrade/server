package app

// Main entrypoint for the cluster node. This must be called once and only
// once, and as the last call in the Go main() function. There are no return
// values as all application operations, logging, and I/O will be forever
// transferred.
func Main(conf Config) error {
	confInit(&conf)
	conf.AddService(redisService())

	hclogger := logInit(conf)
	tm := remoteTimeInit(conf)
	dir, data := dataDirInit(conf)
	m := machineInit(conf, dir, data)
	tlscfg := tlsInit(conf)
	svr, addr := serverInit(conf, tlscfg)
	trans := transportInit(conf, tlscfg, svr, hclogger)
	lstore, sstore := storeInit(conf, dir)
	snaps := snapshotInit(conf, dir, m, hclogger)
	ra := raftInit(conf, hclogger, m, lstore, sstore, snaps, trans)

	joinClusterIfNeeded(conf, ra, addr, tlscfg)
	startUserServices(conf, svr, m, ra)

	//_ = tm
	go runMaintainServers(ra)
	go runWriteApplier(conf, m, ra)
	go runLogLoadedPoller(conf, m, ra, tlscfg)
	go runTicker(conf, tm, m, ra)

	return svr.serve()
}
