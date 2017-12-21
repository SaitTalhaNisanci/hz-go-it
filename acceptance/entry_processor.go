package acceptance

import (
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"log"
)

type EntryProcessor struct {
	classId           int32
	value             string
	identifiedFactory *IdentifiedFactory
	count             int
}

func CreateEntryProcessor(value string) *EntryProcessor {
	processor := &EntryProcessor{classId: 1, value: value}
	identifiedFactory := &IdentifiedFactory{factoryId: 66, entryProcessor: processor}
	processor.identifiedFactory = identifiedFactory
	processor.count = 0
	return processor
}

type IdentifiedFactory struct {
	entryProcessor *EntryProcessor
	factoryId      int32
}

func (identifiedFactory *IdentifiedFactory) Create(id int32) serialization.IdentifiedDataSerializable {
	if id == identifiedFactory.entryProcessor.classId {
		return &EntryProcessor{classId: 1}
	} else {
		return nil
	}
}

func (entryProcessor *EntryProcessor) ReadData(input serialization.DataInput) error {
	var err error
	entryProcessor.value, err = input.ReadUTF()
	log.Println("Read : " + entryProcessor.value)
	return err
}

func (entryProcessor *EntryProcessor) WriteData(output serialization.DataOutput) error {
	log.Println("Write: " + entryProcessor.value)
	entryProcessor.count += 1
	output.WriteUTF(entryProcessor.value)
	return nil
}

func (entryProcessor *EntryProcessor) FactoryId() int32 {
	return entryProcessor.identifiedFactory.factoryId
}

func (entryProcessor *EntryProcessor) ClassId() int32 {
	return entryProcessor.classId
}
