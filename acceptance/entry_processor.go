package acceptance

import (
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"log"
)

type EntryProcessor struct {
	classID           int32
	value             string
	identifiedFactory *IdentifiedFactory
	count             int
}

func CreateEntryProcessor(value string) *EntryProcessor {
	processor := &EntryProcessor{classID: 1, value: value}
	identifiedFactory := &IdentifiedFactory{factoryID: 66, entryProcessor: processor}
	processor.identifiedFactory = identifiedFactory
	processor.count = 0
	return processor
}

type IdentifiedFactory struct {
	entryProcessor *EntryProcessor
	factoryID      int32
}

func (identifiedFactory *IdentifiedFactory) Create(id int32) serialization.IdentifiedDataSerializable {
	if id == identifiedFactory.entryProcessor.classID {
		return &EntryProcessor{classID: 1}
	} else {
		return nil
	}
}

func (entryProcessor *EntryProcessor) ReadData(input serialization.DataInput) error {
	entryProcessor.value = input.ReadUTF()
	log.Println("Read : " + entryProcessor.value)
	return input.Error()
}

func (entryProcessor *EntryProcessor) WriteData(output serialization.DataOutput) error {
	log.Println("Write: " + entryProcessor.value)
	entryProcessor.count += 1
	output.WriteUTF(entryProcessor.value)
	return nil
}

func (entryProcessor *EntryProcessor) FactoryID() int32 {
	return entryProcessor.identifiedFactory.factoryID
}

func (entryProcessor *EntryProcessor) ClassID() int32 {
	return entryProcessor.classID
}
