/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package extractor

import (
	"encoding/json"
	"fmt"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
)

// 接收来自kafka消息,解析成不同数据类型后使用lib/eventemitter模块发送

type ExtractorService struct {
	consumer *kafka.ConsumerRegister
}

const (
	kafka_topic = kafka.Kafka_Topic_Extractor_EventOnChain
)

func Initialize(options kafka.KafkaOptions, group string) error {
	var serv ExtractorService

	serv.consumer = &kafka.ConsumerRegister{}
	serv.consumer.Initialize(options.Brokers)
	if err := serv.consumer.RegisterTopicAndHandler(kafka_topic, group, types.KafkaOnChainEvent{}, serv.handle); err != nil {
		return err
	}

	return nil
}

func (s *ExtractorService) handle(input interface{}) error {
	src, ok := input.(*types.KafkaOnChainEvent)
	if !ok {
		return fmt.Errorf("extractor,input type should be *KafkaOnChainEvent")
	}

	event, err := Disassemble(src)
	if err != nil {
		return fmt.Errorf("extractor, disassemble error:%s", err.Error())
	}

	// todo: delete after test
	log.Debugf("extractor, consume topic:%s ,data:%s", src.Topic, src.Data)

	eventemitter.Emit(src.Topic, event)

	return nil
}

// convert types.kafkaOnChainEvent to kind of events
func Disassemble(src *types.KafkaOnChainEvent) (interface{}, error) {
	event := topicToEvent(src.Topic)
	if event == nil {
		return nil, fmt.Errorf("get event from topic error:cann't found any match topic")
	}

	if err := json.Unmarshal([]byte(src.Data), event); err != nil {
		return nil, err
	}

	return event, nil
}

func Assemble(input interface{}) (*types.KafkaOnChainEvent, error) {
	topic := eventToTopic(input)
	if topic == "" {
		return nil, fmt.Errorf("get topic from event error:cann't found any match topic")
	}

	// marshal event to bytes
	bs, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	// set kafkaOnChainEvent
	dst := &types.KafkaOnChainEvent{}
	dst.Data = string(bs)
	dst.Topic = topic

	return dst, nil
}

func topicToEvent(topic string) interface{} {
	var event interface{}

	switch topic {
	case eventemitter.Miner_SubmitRing_Method:
		event = &types.SubmitRingMethodEvent{}
	case eventemitter.CancelOrder:
		event = &types.OrderCancelledEvent{}
	case eventemitter.CutoffAll:
		event = &types.CutoffEvent{}
	case eventemitter.CutoffPair:
		event = &types.CutoffPairEvent{}
	case eventemitter.Approve:
		event = &types.ApprovalEvent{}
	case eventemitter.Transfer:
		event = &types.TransferEvent{}
	case eventemitter.WethDeposit:
		event = &types.WethDepositEvent{}
	case eventemitter.WethWithdrawal:
		event = &types.WethWithdrawalEvent{}
	case eventemitter.TokenRegistered:
		event = &types.TokenRegisterEvent{}
	case eventemitter.TokenUnRegistered:
		event = &types.TokenUnRegisterEvent{}
	case eventemitter.AddressAuthorized:
		event = &types.AddressAuthorizedEvent{}
	case eventemitter.AddressDeAuthorized:
		event = &types.AddressDeAuthorizedEvent{}

	case eventemitter.RingMined:
		event = &types.RingMinedEvent{}
	case eventemitter.OrderFilled:
		event = &types.OrderFilledEvent{}
	case eventemitter.EthTransfer:
		event = &types.EthTransferEvent{}
	case eventemitter.UnsupportedContract:
		event = &types.UnsupportedContractEvent{}

	case eventemitter.Block_New:
		event = &types.BlockEvent{}
	case eventemitter.Block_End:
		event = &types.BlockEvent{}
	case eventemitter.SyncChainComplete:
		event = &types.SyncCompleteEvent{}
	case eventemitter.ChainForkDetected:
		event = &types.ForkedEvent{}

	default:
		event = nil
	}

	return event
}

func eventToTopic(event interface{}) string {
	var topic string

	switch e := event.(type) {
	case *types.SubmitRingMethodEvent:
		topic = eventemitter.Miner_SubmitRing_Method
	case *types.OrderCancelledEvent:
		topic = eventemitter.CancelOrder
	case *types.CutoffEvent:
		topic = eventemitter.CutoffAll
	case *types.CutoffPairEvent:
		topic = eventemitter.CutoffPair
	case *types.ApprovalEvent:
		topic = eventemitter.Approve
	case *types.TransferEvent:
		topic = eventemitter.Transfer
	case *types.WethDepositEvent:
		topic = eventemitter.WethDeposit
	case *types.WethWithdrawalEvent:
		topic = eventemitter.WethWithdrawal
	case *types.TokenRegisterEvent:
		topic = eventemitter.TokenRegistered
	case *types.TokenUnRegisterEvent:
		topic = eventemitter.TokenUnRegistered
	case *types.AddressAuthorizedEvent:
		topic = eventemitter.AddressAuthorized
	case *types.AddressDeAuthorizedEvent:
		topic = eventemitter.AddressDeAuthorized

	case *types.RingMinedEvent:
		topic = eventemitter.RingMined
	case *types.OrderFilledEvent:
		topic = eventemitter.OrderFilled
	case *types.EthTransferEvent:
		topic = eventemitter.EthTransfer
	case *types.UnsupportedContractEvent:
		topic = eventemitter.UnsupportedContract

	case *types.BlockEvent:
		if e.IsFinished {
			topic = eventemitter.Block_End
		} else {
			topic = eventemitter.Block_New
		}
	case *types.SyncCompleteEvent:
		topic = eventemitter.SyncChainComplete
	case *types.ForkedEvent:
		topic = eventemitter.ChainForkDetected

	default:
		topic = ""
	}

	return topic
}
