package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	NUM_CONCURRENT = 50 // 同時に処理するファイル数
)

func cwalk(flags *flagBundle) error {
	errs := make(chan error, NUM_CONCURRENT)
	lock := make(chan struct{}, NUM_CONCURRENT)
	passedAll := make(chan struct{})
	stopWalking, totalErr := handleProcesses(flags.debug, errs, lock, passedAll)

	// walk を抜けるのは, ↓の2通り
	// 1. walk 中にエラーが発生しなかった -> totalErr に nil が送信されている
	// 2. walk 中にエラーが発生した -> totalErr にエラーが送信されている
	err := filepath.Walk(flags.src, func(path string, info fs.FileInfo, err error) error {
		rpath, err := filepath.Rel(flags.src, path)
		if err != nil {
			return err
		}
		newpath := filepath.Join(flags.dst, rpath)
		if info.IsDir() {
			if _, err := os.Stat(newpath); !os.IsNotExist(err) {
				return nil
			}
			if err := os.Mkdir(newpath, 0o777); err == nil {
				return nil
			} else {
				return err
			}
		}

		if filepath.Ext(path) != ".md" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			newfile, err := os.Create(newpath)
			if err != nil {
				return err
			}
			defer newfile.Close()
			io.Copy(newfile, file)
			return nil
		}

		select {
		case <-stopWalking:
			return filepath.SkipDir
		case lock <- struct{}{}:
			go func() {
				err := process(flags.src, path, newpath, flags)
				if err == nil {
					errs <- nil
					return
				}

				public, debug := handleErr(path, err)
				if public == nil && debug == nil {
					errs <- nil
					return
				}

				if flags.debug {
					errs <- debug
				} else {
					errs <- public
				}
			}()
		}
		return nil
	})

	// walk の終了を handleProcesses に伝える
	close(passedAll)

	if err != nil {
		return err
	}

	return <-totalErr
}

// 正常終了 -> senderr に nil を返す
// 異常終了 -> senderr に エラーを格納 & stop チャネルを閉じる
func handleProcesses(debugmode bool, geterr <-chan error, lock chan struct{}, passedAll <-chan struct{}) (stopWalking chan struct{}, senderr chan error) {
	senderr = make(chan error)
	stopWalking = make(chan struct{})
	go func() {
		for {
			select {
			case <-lock:
			case <-passedAll: // すべてのディレクトリの walk が終了したら, return
				senderr <- nil
				return
			}

			err := <-geterr
			if err != nil {
				close(stopWalking) // エラーをチャネルに流すより先に close しておかないと, ブロックしてしまう
				senderr <- err
				return
			}
		}
	}()
	return stopWalking, senderr
}
