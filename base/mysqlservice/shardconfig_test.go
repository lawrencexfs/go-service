package mysqlservice

// func TestInitMysqlShardCfg(t *testing.T) {
// 	shardCfg, err := GetShardConfig("")
// 	if shardCfg == nil || err != nil {
// 		t.Fatalf("unable to decode into struct, %v", err)
// 	}

// 	fmt.Println("====")

// 	bRet := CheckConfigValidate(shardCfg)
// 	if !bRet {
// 		t.Fatalf("verfity struct failed, %v", err)
// 	}

// }

// func TestGetPressCfg(t *testing.T) {
// 	shardCfg, err := GetShardConfig("")
// 	if shardCfg == nil || err != nil {
// 		t.Fatalf("unable to decode into struct, %v", err)
// 	}

// 	errCount := 0
// 	for i := 0; i < 1000; i++ {
// 		_, err := GetCellCfgByUID(uint64(i))
// 		if err == nil {
// 			// log.Info("cfg.add:", cellCfg.Addr, ",i:", i)
// 		} else {
// 			errCount++
// 		}
// 	}
// 	fmt.Println(errCount)
// }

// func TestGetCellCfg(t *testing.T) {
// 	shardCfg, err := GetShardConfig("")
// 	if shardCfg == nil || err != nil {
// 		t.Fatalf("unable to decode into struct, %v", err)
// 	}

// 	cellCfg, err := GetCellCfgByUID(10)
// 	if err != nil {
// 		// t.Fatal("GetCellCfgByUID failed")
// 		fmt.Errorf("GetCellCfgByUID failed")
// 		return
// 	}
// 	log.Info("cfg.add:", cellCfg.Addr)
// 	wg := sync.WaitGroup{}

// 	wg.Add(3)
// 	go func() {
// 		errCount := 0
// 		for i := 0; i < 50000; {
// 			_, err := GetCellCfgByUID(uint64(i))
// 			if err == nil {
// 				// log.Info("cfg.add:", cellCfg.Addr, ",i:", i)
// 			} else {
// 				errCount++
// 			}
// 			i += 2
// 		}
// 		fmt.Println(errCount)
// 		wg.Done()
// 	}()

// 	go func() {
// 		errCount := 0
// 		for i := 0; i < 50000; {
// 			_, err := GetCellCfgByUID(uint64(i))
// 			if err == nil {
// 				// log.Info("cfg.add:", cellCfg.Addr, ",i:", i)
// 			} else {
// 				errCount++
// 			}

// 			i += 3
// 		}
// 		fmt.Println(errCount)
// 		wg.Done()
// 	}()

// 	go func() {
// 		errCount := 0
// 		for i := 0; i < 50000; {
// 			_, err := GetCellCfgByUID(uint64(i))
// 			if err == nil {
// 				// log.Info("cfg.add:", cellCfg.Addr, ",i:", i)
// 			} else {
// 				errCount++
// 			}
// 			i += 4
// 		}
// 		fmt.Println(errCount)
// 		wg.Done()
// 	}()

// 	wg.Wait()

// 	// 运行总时间 0.063s
// }
