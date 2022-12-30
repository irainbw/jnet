package log

import (
	"fmt"
	"io/fs"
	"jnet/log/fileutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type FileWriter struct {
	closed      int32         // writer是否已关闭
	filePath    string        // 文件保存路径
	fileName    string        // 文件名
	maxSize     int64         // 文件大小上限
	maxSaveTime time.Duration // 文件保存最长时间
	currentSize int64         // 当前文件大小（记录当前已经写入的字节数）
	file        *os.File      // 文件句柄
	writeBuffer chan []byte   // 写缓存
	levels      []Level
	exit        chan struct{}
	WgWrapper
	sync.Once
}

func NewFileWriter(filePath, fileName string, maxSize int64, maxSaveTime time.Duration) (*FileWriter, error) {
	if err := fileutil.NewPath(filePath); err != nil {
		return nil, err
	}
	name := path.Join(filePath, fileName)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	f := &FileWriter{
		filePath:    filePath,
		fileName:    fileName,
		maxSize:     maxSize,
		maxSaveTime: maxSaveTime,
		file:        file,
		writeBuffer: make(chan []byte, 1<<10),
		levels:      AllLevels,
		currentSize: stat.Size(),
		exit:        make(chan struct{}),
	}
	f.Start()
	return f, nil
}

func (f *FileWriter) SetLevels(levels []Level) {
	f.levels = levels
}

func (f *FileWriter) Start() {
	f.WgWrapper.Wrap(func() {
		var ticker *time.Ticker
		var duration = f.checkLife()
		if duration > 0 {
			ticker = time.NewTicker(duration)
		}
		for {
			select {
			case content := <-f.writeBuffer:
				f.writeToFile(content)
				f.rotate()
			case <-f.exit:
				if ticker != nil {
					ticker.Stop()
				}
				return
			}
			if ticker != nil {
				select {
				case <-ticker.C:
					duration = f.checkLife()
					ticker.Reset(duration)
				default:
					break
				}
			}
		}
	})
}

func (f *FileWriter) rotate() {
	if f.maxSize <= 0 || f.currentSize < f.maxSize {
		return
	}

	_ = f.file.Sync()

	_ = f.file.Close()

	// 重命名
	nowTime := time.Now()
	oldName := f.file.Name()
	newFileName := fmt.Sprintf("%s.log", nowTime.Format("2006_01_02_15_04_05"))
	newName := path.Join(f.filePath, newFileName)
	err := os.Rename(oldName, newName)
	if err != nil {
		fmt.Sprintln(err)
	}

	// 重置size和句柄
	f.currentSize = 0
	f.file, _ = os.OpenFile(oldName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}
func (f *FileWriter) writeToFile(msg ...[]byte) {
	for _, m := range msg {
		n, _ := f.file.Write(m)
		f.currentSize += int64(n)
	}
}

func (f *FileWriter) checkLife() time.Duration {
	if f.maxSaveTime <= 0 {
		return 0
	}
	const duration = 30 * time.Second
	_ = filepath.Walk(f.filePath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".log" || info.Name() == f.fileName {
			return nil
		}

		if info.Size() < f.maxSize {
			return nil
		}
		if time.Now().Sub(info.ModTime()) < f.maxSaveTime {
			return nil
		}
		return os.Remove(path)
	})
	return duration
}

func (f *FileWriter) Levels() []Level {
	return f.levels
}

func (f *FileWriter) LogWrite(b []byte) error {
	if atomic.LoadInt32(&f.closed) == 1 {
		return nil
	}
	f.writeBuffer <- b
	return nil
}
func (f *FileWriter) Close() {
	f.Do(func() {
		atomic.StoreInt32(&f.closed, 1)
		f.exit <- struct{}{}
		f.Wait()
	})
}
