package uploads

import "syscall"

const diskReserveBytes = 512 << 20

func freeBytes(dir string) (int64, bool) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, false
	}
	return int64(stat.Bavail) * int64(stat.Bsize), true
}

func fitsOnDisk(dir string, size int64) bool {
	free, known := freeBytes(dir)
	return !known || free-diskReserveBytes >= size
}
