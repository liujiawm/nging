//Do not edit this file, which is automatically generated by the generator.
package dbschema

import (
	"github.com/webx-top/db"
	"github.com/webx-top/db/lib/factory"
	
	"time"
)

type CollectorExport struct {
	param   *factory.Param
	trans	*factory.Transaction
	objects []*CollectorExport
	
	Id                	uint    	`db:"id,omitempty,pk" bson:"id,omitempty" comment:"" json:"id" xml:"id"`
	PageId            	uint    	`db:"page_id" bson:"page_id" comment:"页面ID" json:"page_id" xml:"page_id"`
	Mapping           	string  	`db:"mapping" bson:"mapping" comment:"字段映射" json:"mapping" xml:"mapping"`
	DestDbAccountId   	string  	`db:"dest_db_account_id" bson:"dest_db_account_id" comment:"目标数据库账号设置" json:"dest_db_account_id" xml:"dest_db_account_id"`
	Created           	uint    	`db:"created" bson:"created" comment:"创建时间" json:"created" xml:"created"`
	Exported          	uint    	`db:"exported" bson:"exported" comment:"最近导出时间" json:"exported" xml:"exported"`
}

func (this *CollectorExport) Trans() *factory.Transaction {
	return this.trans
}

func (this *CollectorExport) Use(trans *factory.Transaction) factory.Model {
	this.trans = trans
	return this
}

func (this *CollectorExport) Objects() []*CollectorExport {
	if this.objects == nil {
		return nil
	}
	return this.objects[:]
}

func (this *CollectorExport) NewObjects() *[]*CollectorExport {
	this.objects = []*CollectorExport{}
	return &this.objects
}

func (this *CollectorExport) NewParam() *factory.Param {
	return factory.NewParam(factory.DefaultFactory).SetTrans(this.trans).SetCollection("collector_export").SetModel(this)
}

func (this *CollectorExport) SetParam(param *factory.Param) factory.Model {
	this.param = param
	return this
}

func (this *CollectorExport) Param() *factory.Param {
	if this.param == nil {
		return this.NewParam()
	}
	return this.param
}

func (this *CollectorExport) Get(mw func(db.Result) db.Result, args ...interface{}) error {
	return this.Param().SetArgs(args...).SetRecv(this).SetMiddleware(mw).One()
}

func (this *CollectorExport) List(recv interface{}, mw func(db.Result) db.Result, page, size int, args ...interface{}) (func() int64, error) {
	if recv == nil {
		recv = this.NewObjects()
	}
	return this.Param().SetArgs(args...).SetPage(page).SetSize(size).SetRecv(recv).SetMiddleware(mw).List()
}

func (this *CollectorExport) ListByOffset(recv interface{}, mw func(db.Result) db.Result, offset, size int, args ...interface{}) (func() int64, error) {
	if recv == nil {
		recv = this.NewObjects()
	}
	return this.Param().SetArgs(args...).SetOffset(offset).SetSize(size).SetRecv(recv).SetMiddleware(mw).List()
}

func (this *CollectorExport) Add() (pk interface{}, err error) {
	this.Created = uint(time.Now().Unix())
	this.Id = 0
	pk, err = this.Param().SetSend(this).Insert()
	if err == nil && pk != nil {
		if v, y := pk.(uint); y {
			this.Id = v
		} else if v, y := pk.(int64); y {
			this.Id = uint(v)
		}
	}
	return
}

func (this *CollectorExport) Edit(mw func(db.Result) db.Result, args ...interface{}) error {
	
	return this.Param().SetArgs(args...).SetSend(this).SetMiddleware(mw).Update()
}

func (this *CollectorExport) Upsert(mw func(db.Result) db.Result, args ...interface{}) (pk interface{}, err error) {
	pk, err = this.Param().SetArgs(args...).SetSend(this).SetMiddleware(mw).Upsert(func(){
		
	},func(){
		this.Created = uint(time.Now().Unix())
	this.Id = 0
	})
	if err == nil && pk != nil {
		if v, y := pk.(uint); y {
			this.Id = v
		} else if v, y := pk.(int64); y {
			this.Id = uint(v)
		}
	}
	return 
}

func (this *CollectorExport) Delete(mw func(db.Result) db.Result, args ...interface{}) error {
	
	return this.Param().SetArgs(args...).SetMiddleware(mw).Delete()
}

func (this *CollectorExport) Count(mw func(db.Result) db.Result, args ...interface{}) (int64, error) {
	return this.Param().SetArgs(args...).SetMiddleware(mw).Count()
}
