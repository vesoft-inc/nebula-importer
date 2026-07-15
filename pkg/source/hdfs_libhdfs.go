//go:build cgo && libhdfs

package source

/*
#cgo LDFLAGS: -lhdfs
#include <errno.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static int libhdfs_errno(void) { return errno; }

typedef int32_t tSize;
typedef time_t tTime;
typedef int64_t tOffset;
typedef uint16_t tPort;

typedef enum tObjectKind {
	kObjectKindFile = 'F',
	kObjectKindDirectory = 'D',
} tObjectKind;

struct hdfsBuilder;
typedef struct hdfs_internal* hdfsFS;
typedef struct hdfsFile_internal* hdfsFile;

typedef struct {
	tObjectKind mKind;
	char *mName;
	tTime mLastMod;
	tOffset mSize;
	short mReplication;
	tOffset mBlockSize;
	char *mOwner;
	char *mGroup;
	short mPermissions;
	tTime mLastAccess;
} hdfsFileInfo;

struct hdfsBuilder *hdfsNewBuilder(void);
void hdfsBuilderSetNameNode(struct hdfsBuilder *bld, const char *nn);
void hdfsBuilderSetUserName(struct hdfsBuilder *bld, const char *userName);
void hdfsBuilderSetKerbTicketCachePath(struct hdfsBuilder *bld, const char *kerbTicketCachePath);
int hdfsBuilderConfSetStr(struct hdfsBuilder *bld, const char *key, const char *val);
hdfsFS hdfsBuilderConnect(struct hdfsBuilder *bld);
int hdfsDisconnect(hdfsFS fs);
hdfsFile hdfsOpenFile(hdfsFS fs, const char* path, int flags, int bufferSize, short replication, tSize blocksize);
int hdfsCloseFile(hdfsFS fs, hdfsFile file);
tSize hdfsRead(hdfsFS fs, hdfsFile file, void* buffer, tSize length);
hdfsFileInfo *hdfsListDirectory(hdfsFS fs, const char* path, int *numEntries);
hdfsFileInfo *hdfsGetPathInfo(hdfsFS fs, const char* path);
void hdfsFreeFileInfo(hdfsFileInfo *hdfsFileInfo, int numEntries);
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
	"unsafe"
)

type libhdfsSource struct {
	c    *Config
	fs   C.hdfsFS
	file C.hdfsFile
	info *libhdfsFileInfo
}

type libhdfsFileInfo struct {
	name       string
	size       int64
	mode       os.FileMode
	modTime    time.Time
	accessTime time.Time
	owner      string
	group      string
	isDir      bool
}

func newLibHDFSSource(c *Config) Source {
	return &libhdfsSource{c: c}
}

func (s *libhdfsSource) Name() string {
	return s.c.HDFS.String()
}

func (s *libhdfsSource) Open() error {
	if err := s.Connect(); err != nil {
		return err
	}

	info, err := s.getPathInfo(s.c.HDFS.Path)
	if err != nil {
		return err
	}

	file, err := s.openFile(s.c.HDFS.Path)
	if err != nil {
		return err
	}

	s.info = info
	s.file = file
	return nil
}

func (s *libhdfsSource) Glob() ([]*Config, error) {
	matches, err := sourceGlob(s, s.c.HDFS.Path)
	if err != nil {
		return nil, err
	}

	cs := make([]*Config, 0, len(matches))
	for _, match := range matches {
		cpy := s.c.Clone()
		cpy.HDFS.Path = match
		cs = append(cs, cpy)
	}
	return cs, nil
}

func (s *libhdfsSource) Config() *Config {
	return s.c
}

func (s *libhdfsSource) Size() (int64, error) {
	if s.info == nil {
		return 0, fmt.Errorf("hdfs file %s is not open", s.c.HDFS.Path)
	}

	return s.info.size, nil
}

func (s *libhdfsSource) Read(p []byte) (int, error) {
	if s.file == nil {
		return 0, fmt.Errorf("hdfs file %s is not open", s.c.HDFS.Path)
	}
	if len(p) == 0 {
		return 0, nil
	}

	n := C.hdfsRead(s.fs, s.file, unsafe.Pointer(&p[0]), C.tSize(len(p)))
	switch {
	case n < 0:
		return 0, s.lastError("hdfsRead")
	case n == 0:
		return 0, io.EOF
	default:
		return int(n), nil
	}
}

func (s *libhdfsSource) Close() error {
	var err error
	if s.file != nil {
		if rc := C.hdfsCloseFile(s.fs, s.file); rc != 0 {
			err = s.lastError("hdfsCloseFile")
		}
		s.file = nil
	}
	if s.fs != nil {
		if rc := C.hdfsDisconnect(s.fs); rc != 0 && err == nil {
			err = s.lastError("hdfsDisconnect")
		}
		s.fs = nil
	}
	s.info = nil
	return err
}

func (s *libhdfsSource) IsDir(dir string) (bool, error) {
	if err := s.Connect(); err != nil {
		return false, err
	}

	info, err := s.getPathInfo(dir)
	if err != nil {
		return false, err
	}
	return info.isDir, nil
}

func (s *libhdfsSource) Readdirnames(dir string) ([]string, error) {
	if err := s.Connect(); err != nil {
		return nil, err
	}

	cDir := C.CString(dir)
	defer C.free(unsafe.Pointer(cDir))

	var numEntries C.int
	entries := C.hdfsListDirectory(s.fs, cDir, &numEntries)
	if entries == nil {
		if numEntries == 0 && C.libhdfs_errno() == 0 {
			return []string{}, nil
		}
		return nil, s.lastError("hdfsListDirectory")
	}
	defer C.hdfsFreeFileInfo(entries, numEntries)

	count := int(numEntries)
	names := make([]string, 0, count)
	entrySlice := unsafe.Slice(entries, count)
	for i := 0; i < count; i++ {
		names = append(names, path.Base(C.GoString(entrySlice[i].mName)))
	}
	return names, nil
}

func (s *libhdfsSource) Connect() error {
	if s.fs != nil {
		return nil
	}

	if err := prepareLibHDFSEnv(s.c.HDFS); err != nil {
		return err
	}

	builder := C.hdfsNewBuilder()
	if builder == nil {
		return fmt.Errorf("hdfsNewBuilder returned nil")
	}

	nameNode, err := normalizeLibHDFSNameNode(s.c.HDFS.Address)
	if err != nil {
		return err
	}
	cNameNode := C.CString(nameNode)
	defer C.free(unsafe.Pointer(cNameNode))
	C.hdfsBuilderSetNameNode(builder, cNameNode)

	if s.c.HDFS.User != "" {
		cUser := C.CString(s.c.HDFS.User)
		defer C.free(unsafe.Pointer(cUser))
		C.hdfsBuilderSetUserName(builder, cUser)
	}

	if s.c.HDFS.CCacheFile != "" {
		ccacheFile := normalizedCCacheFile(s.c.HDFS.CCacheFile)
		cCache := C.CString(ccacheFile)
		defer C.free(unsafe.Pointer(cCache))
		C.hdfsBuilderSetKerbTicketCachePath(builder, cCache)
	}

	if s.c.HDFS.ServicePrincipalName != "" {
		if err = hdfsBuilderConfSetStr(builder, "dfs.namenode.kerberos.principal", s.c.HDFS.ServicePrincipalName); err != nil {
			return err
		}
	}

	if s.c.HDFS.DataTransferProtection != "" {
		if err = hdfsBuilderConfSetStr(builder, "dfs.data.transfer.protection", s.c.HDFS.DataTransferProtection); err != nil {
			return err
		}
	}

	fs := C.hdfsBuilderConnect(builder)
	if fs == nil {
		return s.lastError("hdfsBuilderConnect")
	}
	s.fs = fs
	return nil
}

func (s *libhdfsSource) getPathInfo(target string) (*libhdfsFileInfo, error) {
	cPath := C.CString(target)
	defer C.free(unsafe.Pointer(cPath))

	info := C.hdfsGetPathInfo(s.fs, cPath)
	if info == nil {
		return nil, s.lastError("hdfsGetPathInfo")
	}
	defer C.hdfsFreeFileInfo(info, 1)

	return newLibHDFSFileInfo(*info), nil
}

func (s *libhdfsSource) openFile(target string) (C.hdfsFile, error) {
	cPath := C.CString(target)
	defer C.free(unsafe.Pointer(cPath))

	file := C.hdfsOpenFile(s.fs, cPath, C.O_RDONLY, 0, 0, 0)
	if file == nil {
		return nil, s.lastError("hdfsOpenFile")
	}
	return file, nil
}

func (s *libhdfsSource) lastError(op string) error {
	if errno := C.libhdfs_errno(); errno != 0 {
		return &os.PathError{Op: op, Path: s.c.HDFS.Path, Err: fmt.Errorf("%s", C.strerror(errno))}
	}
	return &os.PathError{Op: op, Path: s.c.HDFS.Path, Err: os.ErrNotExist}
}

func newLibHDFSFileInfo(info C.hdfsFileInfo) *libhdfsFileInfo {
	isDir := info.mKind == C.kObjectKindDirectory
	mode := os.FileMode(info.mPermissions)
	if isDir {
		mode |= os.ModeDir
	}

	return &libhdfsFileInfo{
		name:       path.Base(C.GoString(info.mName)),
		size:       int64(info.mSize),
		mode:       mode,
		modTime:    time.Unix(int64(info.mLastMod), 0),
		accessTime: time.Unix(int64(info.mLastAccess), 0),
		owner:      C.GoString(info.mOwner),
		group:      C.GoString(info.mGroup),
		isDir:      isDir,
	}
}

func normalizeLibHDFSNameNode(address string) (string, error) {
	switch {
	case address == "":
		return "default", nil
	case strings.Contains(address, ","):
		return "", fmt.Errorf("the libhdfs backend requires a single namenode address or a nameservice URI, got %q", address)
	case strings.Contains(address, "://"):
		return address, nil
	case strings.Contains(address, ":"):
		return "hdfs://" + address, nil
	default:
		return address, nil
	}
}

func normalizedCCacheFile(path string) string {
	if strings.HasPrefix(path, "FILE:") {
		return strings.TrimPrefix(path, "FILE:")
	}
	return path
}

func hdfsBuilderConfSetStr(builder *C.struct_hdfsBuilder, key, value string) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	if rc := C.hdfsBuilderConfSetStr(builder, cKey, cValue); rc != 0 {
		return fmt.Errorf("failed to set libhdfs config %s: %w", key, os.NewSyscallError("hdfsBuilderConfSetStr", nil))
	}
	return nil
}
