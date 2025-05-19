package storage

import (
	"encoding/binary"
	"fmt"
	"go.uber.org/zap"
	"os"
	. "reportdb/logger"
	. "reportdb/utils"
	"strconv"
	"sync"
	"syscall"
)

type FileHandle struct {
	file *os.File

	availableSize int64

	lastBlockEnd int64

	lock *sync.RWMutex

	mappedBuffer []byte
}

type FileManager struct {
	baseDir string

	fileHandles map[uint8]*FileHandle // fileHandles[partitionId]

	lock *sync.RWMutex
}

func NewFileManager(baseDir string) *FileManager {

	return &FileManager{

		fileHandles: make(map[uint8]*FileHandle),

		lock: &sync.RWMutex{},

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

	file, err := os.OpenFile(partitionFile, os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {

		return nil, err
	}

	fileInfo, err := file.Stat()

	if err != nil {

		err := file.Close()

		return nil, err
	}

	handle = &FileHandle{

		file: file,

		availableSize: fileInfo.Size(),

		lock: &sync.RWMutex{},
	}

	mappedBuffer, err := syscall.Mmap(int(handle.file.Fd()), 0, int(handle.availableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil && handle.availableSize != 0 {

		return nil, fmt.Errorf("mmap error: %v", err)
	}

	if handle.availableSize > 0 {

		handle.lastBlockEnd = int64(binary.LittleEndian.Uint64(mappedBuffer[:8]))

	} else {

		handle.lastBlockEnd = handle.availableSize + 8
	}

	handle.mappedBuffer = mappedBuffer

	fileManager.fileHandles[partition] = handle

	return handle, nil
}

func (fileManager *FileManager) CheckCapacity(handle *FileHandle, entryList []*IndexEntry, requiredSize int64) ([]*IndexEntry, error) {

	handle.lock.Lock()

	defer handle.lock.Unlock()

	if len(entryList) > 0 {

		lastEntry := entryList[len(entryList)-1]

		if lastEntry.EntryEnd+requiredSize <= lastEntry.BlockEnd {

			return entryList, nil
		}
	}

	if handle.mappedBuffer != nil {

		if err := syscall.Munmap(handle.mappedBuffer); err != nil {

			return nil, fmt.Errorf("munmap failed: %v", err)
		}
	}

	fileGrowthSize := int64(GetFileGrowthSize())

	handle.availableSize = handle.lastBlockEnd + fileGrowthSize

	if err := handle.file.Truncate(handle.availableSize); err != nil {

		return nil, fmt.Errorf("failed to grow file: %v", err)
	}

	mappedBuffer, err := syscall.Mmap(int(handle.file.Fd()), 0, int(handle.availableSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		return nil, fmt.Errorf("mmap failed: %v", err)
	}

	handle.mappedBuffer = mappedBuffer

	entryList = append(entryList, &IndexEntry{

		BlockStart: handle.lastBlockEnd,

		BlockEnd: handle.availableSize,

		EntryStart: handle.lastBlockEnd,

		EntryEnd: handle.lastBlockEnd,
	})

	handle.lastBlockEnd = handle.availableSize

	binary.LittleEndian.PutUint64(handle.mappedBuffer[:8], uint64(handle.lastBlockEnd))

	return entryList, nil
}

func (fileManager *FileManager) Close() {

	fileManager.lock.Lock()

	defer fileManager.lock.Unlock()

	for _, handle := range fileManager.fileHandles {

		if handle.mappedBuffer != nil {

			if err := syscall.Munmap(handle.mappedBuffer); err != nil {

				Logger.Error("FileManager: munmap failed during close", zap.Error(err))

			}
		}

		if err := handle.file.Close(); err != nil {

			Logger.Error("FileManager: file close failed", zap.Error(err))
		}

	}

}
