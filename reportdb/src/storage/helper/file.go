package helper

import (
	"fmt"
	"os"
	. "reportdb/utils"
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
	fileHandles map[uint8]*FileHandle // fileHandles[partitionId]

	handleMutex sync.RWMutex

	baseDir string
}

func NewFileManager(baseDir string) *FileManager {

	return &FileManager{

		fileHandles: make(map[uint8]*FileHandle),

		handleMutex: sync.RWMutex{},

		baseDir: baseDir, // ./database/YYYY/MM/DD/counter_1
	}
}

func (fileManager *FileManager) GetHandle(partition uint8) (*FileHandle, error) {

	fileManager.handleMutex.RLock()

	if handle, exists := fileManager.fileHandles[partition]; exists {

		fileManager.handleMutex.RUnlock()

		return handle, nil
	}

	fileManager.handleMutex.RUnlock()

	fileManager.handleMutex.Lock()

	defer fileManager.handleMutex.Unlock()

	partitionFile := fileManager.baseDir + "/partition_" + strconv.Itoa(int(partition)) + ".bin"

	if err := os.MkdirAll(fileManager.baseDir, 0755); err != nil {

		return nil, err
	}

	file, err := os.OpenFile(partitionFile, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {

		return nil, err
	}

	fileInfo, err := file.Stat()

	if err != nil {

		err := file.Close()

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

	fileManager.fileHandles[partition] = handle

	return handle, nil
}

func (fileManager *FileManager) CheckCapacity(handle *FileHandle, requiredSize int64) error {

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

	handle.AvailableSize = handle.Offset + GetFileGrowthSize()

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
