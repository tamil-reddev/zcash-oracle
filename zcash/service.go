// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zcash

import (
	"errors"
	"net/http"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/json"
	ej "encoding/json"
)

var (
	errNoSuchBlock           = errors.New("couldn't get block from database. Does it exist?")
	errCannotGetLastAccepted = errors.New("problem getting last accepted")
	errNoSuchData  = errors.New("No data found!!")
)

// Service is the API service for this VM
type Service struct{ vm *VM }

// GetBlockArgs are the arguments to GetBlock
type GetBlockArgs struct {
	// ID of the block we're getting.
	// If left blank, gets the latest block
	ID *ids.ID `json:"id"`
}

// GetBlockReply is the reply from GetBlock
type GetBlockReply struct {
	Timestamp json.Uint64 `json:"timestamp"` // Timestamp of block
	Data      ZcashBlock  `json:"data"`  // Data of zcash block
	Height    json.Uint64 `json:"height"`    // Height of block
	ID        ids.ID      `json:"id"`        // String repr. of ID of block
	ParentID  ids.ID      `json:"parentID"`  // String repr. of ID of block's parent
}

// GetBlock gets the block whose ID is [args.ID]
// If [args.ID] is empty, get the latest block
func (s *Service) GetBlock(_ *http.Request, args *GetBlockArgs, reply *GetBlockReply) error {
	// If an ID is given, parse its string representation to an ids.ID
	// If no ID is given, ID becomes the ID of last accepted block
	var (
		id  ids.ID
		err error
	)

	if args.ID == nil {
		id, err = s.vm.state.GetLastAccepted()
		if err != nil {
			return errCannotGetLastAccepted
		}
	} else {
		id = *args.ID
	}

	// Get the block from the database
	block, err := s.vm.getBlock(id)
	if err != nil {
		return errNoSuchBlock
	}

	// Fill out the response with the block's data
	assignValues(reply, block)

	return err
}

type QueryDataArgs struct {
	ID uint64 `json:"id"`
}



// GetBlock gets the block whose ID is [args.ID]
// If [args.ID] is empty, get the latest block
func (s *Service) GetBlockByHeight(_ *http.Request, args *QueryDataArgs, reply *GetBlockReply) error {

	var (
		id  uint64
	)

	if args.ID == 0 {
		return errNoSuchBlock
	} else {
		id = args.ID
	}

	block, err := s.vm.getBlockByHeight(id)
	if err != nil {
		fmt.Printf("Error in finding getBlockByHeight : %+v\n", err)
	}

	if block ==  nil {
		// Get the block from the database
		resp, err := s.vm.queryData(id)

		jsonData, err := ej.Marshal(resp)
		
		byteArray := []byte(jsonData)

		if len(byteArray) > 0 {
			status := s.vm.addZcashBlock(byteArray)		
			fmt.Printf("block added into subnet : %+v %+v\n", status, resp.Height)
			if err != nil {
				return errNoSuchBlock
			}
		}
		
		
		return err

	} else {	
		fmt.Printf("block found in subnet : %+v\n", block.Height)
		// Assign values from resp to reply
		assignValues(reply, block)
		return nil
	}	
	
}

func assignValues(reply *GetBlockReply, block *Block) {

	// Fill out the response with the block's data
	reply.Timestamp = json.Uint64(block.Timestamp().Unix())
	data := block.Data()
	if len(data) != 0 {
		zblock := ZcashBlock{}
		ej.Unmarshal(data, &zblock)
		reply.Data = zblock
	}
	reply.Height = json.Uint64(block.Hght)
	reply.ID = block.ID()
	reply.ParentID = block.Parent()
}
