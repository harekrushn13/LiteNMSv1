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

	MappedBuffer []byte
}

type FileManager struct {
	fileHandles map[uint8]*FileHandle // fileHandles[partitionId]

	lock sync.RWMutex

	baseDir string
}

func NewFileManager(baseDir string) *FileManager {

	return &FileManager{

		fileHandles: make(map[uint8]*FileHandle),

		lock: sync.RWMutex{},

		baseDir: baseDir, // ./database/YYYY/MM/DD/counter_1
	}
}

func (fileManager *FileManager) GetHandle(partition uint8) (*FileHandle, error) {

	fileManager.lock.RLock()

	handle, exists := fileManager.fileHandles[partition]

	fileManager.lock.RUnlock()

	if exists {

		return handle, nil
	}

	fileManager.lock.Lock()

	defer fileManager.lock.Unlock()

	if handle, exists = fileManager.fileHandles[partition]; exists {

		return handle, nil
	}

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

	handle = &FileHandle{

		File: file,

		AvailableSize: fileInfo.Size(),

		Lock: &sync.RWMutex{},
	}

	mappedBuffer, err := syscall.Mmap(int(handle.File.Fd()), 0, int(handle.AvailableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if handle.AvailableSize > 0 {

		handle.Offset = fileInfo.Size() - 4

	} else {

		handle.Offset = fileInfo.Size() + 4
	}

	handle.MappedBuffer = mappedBuffer

	fileManager.fileHandles[partition] = handle

	return handle, nil
}

func (fileManager *FileManager) CheckCapacity(handle *FileHandle, requiredSize int64) error {

	handle.Lock.Lock()

	defer handle.Lock.Unlock()

	if handle.Offset+requiredSize <= handle.AvailableSize {

		return nil
	}

	if handle.MappedBuffer != nil {

		if err := syscall.Munmap(handle.MappedBuffer); err != nil {

			return fmt.Errorf("munmap failed: %v", err)
		}
	}

	fileGrowthSize, err := GetFileGrowthSize()

	if err != nil {

		return err
	}

	handle.AvailableSize = handle.Offset + fileGrowthSize

	if err := handle.File.Truncate(handle.AvailableSize); err != nil {

		return fmt.Errorf("failed to grow file: %v", err)
	}

	mappedBuffer, err := syscall.Mmap(int(handle.File.Fd()), 0, int(handle.AvailableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		return fmt.Errorf("mmap failed: %v", err)
	}

	handle.MappedBuffer = mappedBuffer

	return nil
}
