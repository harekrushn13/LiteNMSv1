package helper

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"syscall"
)

type FileHandle struct {
	File *os.File

	Offset int64

	Lock *sync.RWMutex

	AvailableSize int64

	MmapData []byte
}

type FileManager struct {
	handles map[uint8]*FileHandle // handles [partition]

	handlesMux sync.RWMutex

	BaseDir string

	PartitionCount uint8

	FileGrowth int64 // byte
}

func NewFileManager(baseDir string) *FileManager {

	return &FileManager{

		handles: make(map[uint8]*FileHandle),

		handlesMux: sync.RWMutex{},

		BaseDir: baseDir, // ./src/storage/database/YYYY/MM/DD/counter_1

		FileGrowth: 64,
	}
}

func (fm *FileManager) GetHandle(partition uint8) (*FileHandle, error) {

	fm.handlesMux.RLock()

	if handle, exists := fm.handles[partition]; exists {

		fm.handlesMux.RUnlock()

		return handle, nil
	}

	fm.handlesMux.RUnlock()

	fm.handlesMux.Lock()

	defer fm.handlesMux.Unlock()

	partitionFile := fm.BaseDir + "/partition_" + strconv.Itoa(int(partition)) + ".bin"

	if err := os.MkdirAll(fm.BaseDir, 0755); err != nil {

		return nil, err
	}

	file, err := os.OpenFile(partitionFile, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {

		return nil, err
	}

	fileInfo, err := file.Stat()

	if err != nil {

		file.Close()

		return nil, err
	}

	handle := &FileHandle{

		File: file,

		Offset: fileInfo.Size(),

		AvailableSize: fileInfo.Size(),

		Lock: &sync.RWMutex{},
	}

	data, err := syscall.Mmap(int(handle.File.Fd()), 0, int(handle.AvailableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	handle.MmapData = data

	fm.handles[partition] = handle

	return handle, nil
}

func (fm *FileManager) EnsureCapacity(handle *FileHandle, requiredSize int64) error {

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	if handle.Offset+requiredSize <= handle.AvailableSize {

		return nil
	}

	if handle.MmapData != nil {

		if err := syscall.Munmap(handle.MmapData); err != nil {

			return fmt.Errorf("munmap failed: %v", err)
		}
	}

	handle.AvailableSize = handle.Offset + fm.FileGrowth // just 64 byte adding

	if err := handle.File.Truncate(handle.AvailableSize); err != nil {

		return fmt.Errorf("failed to grow file: %v", err)
	}

	data, err := syscall.Mmap(int(handle.File.Fd()), 0, int(handle.AvailableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		return fmt.Errorf("mmap failed: %v", err)
	}

	handle.MmapData = data

	return nil
}
