package core

import "testing"
import "dedup/sqlite"
func TestDuplicate(t *testing.T) {
	/*
	在这里写一个测试文件，测试Duplicate函数
	*/
	sqlite.SetSqlite()
	Duplicate("D:\\Users\\Public\\Github\\dedup",false)
}
