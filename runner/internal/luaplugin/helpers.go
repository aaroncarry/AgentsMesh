package luaplugin

import (
	lua "github.com/yuin/gopher-lua"
)

// getStringField extracts a string field from a Lua table.
func getStringField(tbl *lua.LTable, key string) string {
	v := tbl.RawGetString(key)
	if lv, ok := v.(lua.LString); ok {
		return string(lv)
	}
	return ""
}

// getIntField extracts an integer field from a Lua table.
func getIntField(tbl *lua.LTable, key string) int {
	v := tbl.RawGetString(key)
	if lv, ok := v.(lua.LNumber); ok {
		return int(lv)
	}
	return 0
}

// getBoolField extracts a boolean field from a Lua table.
func getBoolField(tbl *lua.LTable, key string) bool {
	v := tbl.RawGetString(key)
	if lv, ok := v.(lua.LBool); ok {
		return bool(lv)
	}
	return false
}

// getStringArrayField extracts a string array field from a Lua table.
func getStringArrayField(tbl *lua.LTable, key string) []string {
	v := tbl.RawGetString(key)
	if arrTbl, ok := v.(*lua.LTable); ok {
		var result []string
		arrTbl.ForEach(func(_, val lua.LValue) {
			if s, ok := val.(lua.LString); ok {
				result = append(result, string(s))
			}
		})
		return result
	}
	return nil
}

// parseUIConfig parses UI configuration from a Lua table.
func parseUIConfig(tbl *lua.LTable) *UIConfig {
	ui := &UIConfig{
		Configurable: getBoolField(tbl, "configurable"),
		Fields:       make([]UIField, 0),
	}

	fieldsVal := tbl.RawGetString("fields")
	if fieldsTbl, ok := fieldsVal.(*lua.LTable); ok {
		fieldsTbl.ForEach(func(_, val lua.LValue) {
			if fieldTbl, ok := val.(*lua.LTable); ok {
				field := parseUIField(fieldTbl)
				ui.Fields = append(ui.Fields, field)
			}
		})
	}

	return ui
}

// parseUIField parses a single UI field from a Lua table.
func parseUIField(tbl *lua.LTable) UIField {
	field := UIField{
		Name:        getStringField(tbl, "name"),
		Type:        getStringField(tbl, "type"),
		Label:       getStringField(tbl, "label"),
		Description: getStringField(tbl, "description"),
		Placeholder: getStringField(tbl, "placeholder"),
		Required:    getBoolField(tbl, "required"),
	}

	// Parse default value
	defaultVal := tbl.RawGetString("default")
	switch v := defaultVal.(type) {
	case lua.LString:
		field.Default = string(v)
	case lua.LNumber:
		field.Default = float64(v)
	case lua.LBool:
		field.Default = bool(v)
	}

	// Parse min/max for number fields
	minVal := tbl.RawGetString("min")
	if n, ok := minVal.(lua.LNumber); ok {
		f := float64(n)
		field.Min = &f
	}

	maxVal := tbl.RawGetString("max")
	if n, ok := maxVal.(lua.LNumber); ok {
		f := float64(n)
		field.Max = &f
	}

	// Parse options for select fields
	optionsVal := tbl.RawGetString("options")
	if optionsTbl, ok := optionsVal.(*lua.LTable); ok {
		optionsTbl.ForEach(func(_, val lua.LValue) {
			if optTbl, ok := val.(*lua.LTable); ok {
				opt := UIOption{
					Value: getStringField(optTbl, "value"),
					Label: getStringField(optTbl, "label"),
				}
				field.Options = append(field.Options, opt)
			}
		})
	}

	return field
}

// goValueToLua converts a Go value to a Lua value.
func goValueToLua(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case int:
		return lua.LNumber(val)
	case int64:
		return lua.LNumber(val)
	case float64:
		return lua.LNumber(val)
	case string:
		return lua.LString(val)
	case []string:
		tbl := L.NewTable()
		for i, s := range val {
			tbl.RawSetInt(i+1, lua.LString(s))
		}
		return tbl
	case map[string]string:
		tbl := L.NewTable()
		for k, v := range val {
			tbl.RawSetString(k, lua.LString(v))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.NewTable()
		for k, v := range val {
			tbl.RawSetString(k, goValueToLua(L, v))
		}
		return tbl
	default:
		return lua.LNil
	}
}

// luaTableToGoMap converts a Lua table to a Go map.
func luaTableToGoMap(tbl *lua.LTable) map[string]interface{} {
	result := make(map[string]interface{})
	tbl.ForEach(func(key, val lua.LValue) {
		if k, ok := key.(lua.LString); ok {
			result[string(k)] = luaValueToGo(val)
		}
	})
	return result
}

// luaValueToGo converts a Lua value to a Go value.
func luaValueToGo(v lua.LValue) interface{} {
	switch val := v.(type) {
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case lua.LString:
		return string(val)
	case *lua.LTable:
		return luaTableToGoMap(val)
	default:
		return nil
	}
}
