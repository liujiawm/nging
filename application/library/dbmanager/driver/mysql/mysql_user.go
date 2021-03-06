/*
   Nging is a toolbox for webmasters
   Copyright (C) 2018-present  Wenhui Shen <swh@admpub.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package mysql

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/webx-top/com"
	"github.com/webx-top/echo"
)

func (m *mySQL) dropUser(user string, host string) error {
	if len(host) > 0 {
		user = quoteVal(user) + `@` + quoteVal(host)
	} else {
		user = quoteVal(user) + `@''`
	}
	r := &Result{}
	r.SQL = "DROP USER " + user
	r.Exec(m.newParam())
	m.AddResults(r)
	if r.err != nil {
		return r.err
	}
	r2 := &Result{}
	r2.SQL = "FLUSH PRIVILEGES"
	r2.Exec(m.newParam())
	m.AddResults(r2)
	return r2.err
}

func (m *mySQL) editUser(oldUser string, oldHost string, newUser string, newHost string, oldPasswd string, newPasswd string, isHashed bool) error {
	var user string
	if len(oldHost) > 0 {
		user = quoteVal(oldUser) + `@` + quoteVal(oldHost)
	} else {
		user = quoteVal(oldUser) + `@''`
	}
	if len(newUser) == 0 {
		return errors.New(m.T(`用户名不能为空`))
	}

	oldPass, grants, _, err := m.getUserGrants(oldHost, oldUser)
	if err != nil {
		return err
	}
	if len(oldPasswd) == 0 {
		oldPasswd = oldPass
	}

	r := &Result{}
	newUser = quoteVal(newUser) + `@` + quoteVal(newHost)
	if len(newPasswd) > 0 {
		if !isHashed {
			r.SQL = `SELECT PASSWORD(` + quoteVal(newPasswd) + `)`
			row := m.newParam().SetCollection(r.SQL).QueryRow()
			var v sql.NullString
			err := row.Scan(&v)
			if err != nil {
				return err
			}
			newPasswd = v.String
		}
	} else {
		newPasswd = oldPasswd
	}
	var created bool
	onerror := func(err error) error {
		return err
	}
	if user != newUser {
		if len(newPasswd) == 0 {
			return errors.New(m.T(`密码不能为空。请注意：修改用户名的时候，必须设置密码`))
		}
		r.SQL = `GRANT USAGE ON *.* TO`
		if com.VersionCompare(m.version, `5`) >= 0 {
			r.SQL = `CREATE USER`
		}
		r.SQL += ` ` + newUser + ` IDENTIFIED BY PASSWORD ` + quoteVal(newPasswd)
		created = true
		onerror = func(err error) error {
			r2 := &Result{}
			r2.SQL = "DROP USER " + newUser
			r2.Exec(m.newParam())
			if r2.err != nil {
				m.Echo().Logger().Error(r2.err)
			}
			m.AddResults(r2)
			return err
		}
	} else if len(newPasswd) > 0 && oldPasswd != newPasswd {
		r.SQL = `SET PASSWORD FOR ` + newUser + `=` + quoteVal(newPasswd)
	} else {
		r.SQL = ``
	}
	//panic(r.SQL)
	if len(r.SQL) > 0 {
		r.Exec(m.newParam())
		m.AddResults(r)
		if r.err != nil {
			return r.err
		}
	}

	scopes := m.FormValues(`scopes[]`)
	databases := m.FormValues(`databases[]`)
	tables := m.FormValues(`tables[]`)
	columns := m.FormValues(`columns[]`)
	scopeMaxIndex := len(scopes) - 1
	databaseMaxIndex := len(databases) - 1
	tableMaxIndex := len(tables) - 1
	columnMaxIndex := len(columns) - 1
	objects := m.FormValues(`objects[]`)
	newGrants := map[string]*Grant{}

	mapx := echo.NewMapx(m.Forms())
	mapx = mapx.Get(`grants`)
	logger := m.Echo().Logger()
	//objects: objects[0|1|...]=`*.*|db.*|db.table|db.table.col1,col2`
	for k, v := range objects {
		if k > scopeMaxIndex {
			logger.Debugf(`k > scopeMaxIndex: %v > %v`, k, scopeMaxIndex)
			continue
		}
		if k > databaseMaxIndex {
			logger.Debugf(`k > databaseMaxIndex: %v > %v`, k, databaseMaxIndex)
			continue
		}
		if k > tableMaxIndex {
			logger.Debugf(`k > tableMaxIndex: %v > %v`, k, tableMaxIndex)
			continue
		}
		if k > columnMaxIndex {
			logger.Debugf(`k > columnMaxIndex: %v > %v`, k, columnMaxIndex)
			continue
		}
		if len(scopes[k]) == 0 {
			logger.Debugf(`scopes[%v] is not set`, k)
			continue
		}
		gr := &Grant{
			Scope:    scopes[k],
			Value:    v,
			Database: databases[k],
			Table:    tables[k],
			Columns:  columns[k],
			Settings: map[string]string{},
		}
		v = gr.String()
		if oldGr, ok := newGrants[v]; !ok {
			newGrants[v] = gr
		} else {
			for k, v := range oldGr.Settings {
				gr.Settings[k] = v
			}
		}
		if mapx == nil {
			newGrants[v] = gr
			continue
		}
		mp := mapx.Get(strconv.Itoa(k))
		if mp != nil {
			for group, settings := range mp.Map {
				if settings.Map == nil || !gr.IsValid(group, settings.Map) {
					continue
				}
				for name, m := range settings.Map {
					gr.Settings[name] = m.Value()
				}
			}
		}
	}
	hasURLGrantValue := len(m.Form(`grant`)) > 0
	operations := []*Grant{}
	//newGrants: newGrants[*.*|db.*|db.table|db.table(col1,col2)][DROP|...]=`0|1`
	for object, grant := range newGrants {
		onAndCol := reGrantColumn.FindStringSubmatch(object)
		//fmt.Printf("object: %v matched: %#v\n", object, onAndCol)
		if len(onAndCol) < 3 {
			continue
		}
		grant.Operation = &Operation{
			Grant:   []string{},
			Revoke:  []string{},
			Columns: onAndCol[2],
			On:      onAndCol[1],
			User:    newUser,
			Scope:   grant.Scope,
		}
		if hasURLGrantValue {
			for key, val := range grant.Settings {
				if val != `1` {
					grant.Revoke = append(grant.Revoke, key)
				}
			}
		} else if user == newUser {
			if vals, ok := grants[object]; ok {
				for key := range vals {
					if _, ok := grant.Settings[key]; !ok {
						grant.Revoke = append(grant.Revoke, key)
					}
				}
				for key := range grant.Settings {
					if _, ok := vals[key]; !ok {
						grant.Grant = append(grant.Grant, key)
					}
				}
				delete(grants, object)
			} else {
				for key := range grant.Settings {
					grant.Grant = append(grant.Grant, key)
				}
			}
		} else {
			for key := range grant.Settings {
				grant.Grant = append(grant.Grant, key)
			}
		}
		operations = append(operations, grant)

	}
	if len(oldUser) > 0 && (!hasURLGrantValue && !created) {
		for object, revoke := range grants {
			onAndCol := reGrantColumn.FindStringSubmatch(object)
			if len(onAndCol) < 3 {
				continue
			}
			op := &Operation{
				Grant:   []string{},
				Revoke:  []string{},
				Columns: onAndCol[2],
				On:      onAndCol[1],
				User:    newUser,
			}
			for k := range revoke {
				op.Revoke = append(op.Revoke, k)
			}
			if err := op.Apply(m); err != nil {
				return err
			}
		}
	}
	for _, op := range operations {
		if err := op.Apply(m); err != nil {
			return onerror(err)
		}
	}
	if len(oldUser) > 0 {
		if created {
			r := &Result{}
			r.SQL = "DROP USER " + user
			r.Exec(m.newParam())
			m.AddResults(r)
			if r.err != nil {
				return onerror(err)
			}
		}
	}
	r2 := &Result{}
	r2.SQL = "FLUSH PRIVILEGES"
	r2.Exec(m.newParam())
	m.AddResults(r2)
	if r2.err != nil {
		return onerror(err)
	}
	return nil
}

func (m *mySQL) getUserGrants(host, user string) (string, map[string]map[string]bool, []string, error) {
	r := map[string]map[string]bool{}
	var (
		sortNumber []string
		oldPass    string
		err        error
	)
	if len(host) > 0 {
		sqlStr := "SHOW GRANTS FOR " + quoteVal(user) + "@" + quoteVal(host)
		rows, err := m.newParam().SetCollection(sqlStr).Query()
		if err != nil {
			return oldPass, r, sortNumber, err
		}
		defer rows.Close()
		for rows.Next() {
			var v sql.NullString
			err = rows.Scan(&v)
			if err != nil {
				break
			}
			matchOn := reGrantOn.FindStringSubmatch(v.String)
			if len(matchOn) > 0 {
				matchBrackets := reGrantBrackets.FindAllStringSubmatch(matchOn[1], -1)
				if len(matchBrackets) > 0 {
					for _, val := range matchBrackets {
						if val[1] != `USAGE` {
							k := matchOn[2] + val[2]
							if _, ok := r[k]; !ok {
								r[k] = map[string]bool{}
								sortNumber = append(sortNumber, k)
							}
							if val[1] == `PROXY` {
								r[k]["ALL PRIVILEGES"] = true
							}
							r[k][val[1]] = true
						}
						if reGrantOption.MatchString(v.String) {
							k := matchOn[2] + val[2]
							if _, ok := r[k]; !ok {
								r[k] = map[string]bool{}
								sortNumber = append(sortNumber, k)
							}
							r[k]["GRANT OPTION"] = true
						}
					}
				}
			}
			matchIdent := reGrantIdent.FindStringSubmatch(v.String)
			if len(matchIdent) > 0 {
				oldPass = matchIdent[1]
			}
		}
	} else {
		sqlStr := "SELECT SUBSTRING_INDEX(CURRENT_USER, '@', -1)"
		row := m.newParam().SetCollection(sqlStr).QueryRow()
		var v sql.NullString
		err = row.Scan(&v)
		if err != nil {
			return oldPass, r, sortNumber, err
		}
		m.Request().Form().Set(`host`, v.String)
	}
	var key string
	if len(m.dbName) == 0 || len(r) > 0 {
	} else {
		key = com.AddCSlashes(m.dbName, '%', '_', '\\') + ".*"
	}
	r[key] = map[string]bool{}
	sortNumber = append(sortNumber, key)
	return oldPass, r, sortNumber, err
}
