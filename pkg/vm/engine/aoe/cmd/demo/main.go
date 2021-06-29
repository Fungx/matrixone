package main

import (
	"matrixone/pkg/vm/engine/aoe"
	e "matrixone/pkg/vm/engine/aoe/storage"
	"matrixone/pkg/vm/engine/aoe/storage/db"
	"matrixone/pkg/vm/engine/aoe/storage/dbi"
	md "matrixone/pkg/vm/engine/aoe/storage/metadata"
	"matrixone/pkg/vm/engine/aoe/storage/mock/type/chunk"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	workDir := "/tmp/myDemo"
	os.RemoveAll(workDir)
	colCnt := 16
	metaConf := &md.Configuration{
		Dir:              workDir,
		BlockMaxRows:     10000,
		SegmentMaxBlocks: 10,
	}
	cacheCfg := &e.CacheCfg{
		IndexCapacity:  10000,
		InsertCapacity: metaConf.BlockMaxRows * uint64(colCnt) * 100,
		DataCapacity:   metaConf.BlockMaxRows * metaConf.SegmentMaxBlocks * uint64(colCnt) * 2,
	}
	opts := e.Options{CacheCfg: cacheCfg}
	opts.Meta.Conf = metaConf
	inst, err := db.Open(workDir, &opts)
	if err != nil {
		panic(err)
	}

	tableInfo := md.MockTableInfo(colCnt)
	tabletInfo := aoe.TabletInfo{
		Name:  tableInfo.Name,
		Table: *tableInfo,
	}
	tName := tabletInfo.Name
	_, err = inst.CreateTable(&tabletInfo)
	if err != nil {
		panic(err)
	}
	rows := metaConf.BlockMaxRows / 2
	tblMeta, err := inst.Opts.Meta.Info.ReferenceTableByName(tName)
	ck := chunk.MockBatch(tblMeta.Schema.Types(), rows)
	logIdx := &md.LogIndex{
		ID:       uint64(0),
		Capacity: uint64(ck.Vecs[0].Length()),
	}
	cols := make([]int, 0)
	for i := 0; i < len(tblMeta.Schema.ColDefs); i++ {
		cols = append(cols, i)
	}

	var (
		insertWg sync.WaitGroup
		searchWg sync.WaitGroup
	)
	insertCnt := int(float32(metaConf.SegmentMaxBlocks*10) * (float32(rows) / float32(metaConf.BlockMaxRows)))
	insertWg.Add(1)
	go func() {
		for i := 0; i < insertCnt; i++ {
			err = inst.Append(tName, ck, logIdx)
			if err != nil {
				panic(err)
			}
		}
		insertWg.Done()
	}()
	ctx := dbi.GetSnapshotCtx{ScanAll: true, TableName: tName, Cols: cols}

	doScan := func() {
		ss, err := inst.GetSnapshot(&ctx)
		if err != nil {
			panic(err)
		}
		segIt := ss.NewIt()
		for segIt.Valid() {
			segment := segIt.GetHandle()
			blkIt := segment.NewIt()
			for blkIt.Valid() {
				block := blkIt.GetHandle()
				hh := block.Prefetch()
				vec := hh.GetVector(0)
				log.Infof("vec[1]=%v", vec.GetValue(1))
				hh.Close()
				blkIt.Next()
			}
			blkIt.Close()
			segIt.Next()
		}
		segIt.Close()

		ss.Close()
	}
	searchWg.Add(1)
	go func() {
		for i := 0; i < 0; i++ {
			doScan()
		}
		searchWg.Done()
	}()

	insertWg.Wait()
	searchWg.Wait()
	time.Sleep(time.Duration(500) * time.Millisecond)

	doScan()
	log.Info(inst.MTBufMgr.String())
	log.Info(inst.SSTBufMgr.String())
	inst.Close()
}
