package tg

import (
	"encoding/json"

	"github.com/gotd/td/tg"
	"github.com/tidwall/gjson"
)

func getUserName(user *tg.User) string {
	if username, ok := user.GetUsername(); ok && username != "" {
		return "@" + username
	}

	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}

	return name
}

// cleanJSON removes empty/default fields from JSON
func cleanJSON(data []byte) []byte {
	result := gjson.ParseBytes(data)
	cleaned := cleanValue(result)
	if cleaned == nil {
		return data // Return original if cleaning failed
	}

	cleanedJSON, err := json.Marshal(cleaned)
	if err != nil {
		return data // Return original if marshaling failed
	}

	return cleanedJSON
}

func cleanValue(v gjson.Result) interface{} {
	switch v.Type {
	case gjson.String:
		if v.String() == "" {
			return nil
		}
		return v.String()
	case gjson.Number:
		// return nil
		if v.Int() == 0 && v.Float() == 0 {
			return nil
		}
		return v.Value()
	case gjson.True:
		return nil
		// return true
	case gjson.False:
		return nil
	case gjson.Null:
		return nil
	case gjson.JSON:
		if v.IsArray() {
			arr := make([]interface{}, 0)
			v.ForEach(func(_, item gjson.Result) bool {
				if cleaned := cleanValue(item); cleaned != nil {
					arr = append(arr, cleaned)
				}
				return true
			})
			if len(arr) == 0 {
				return nil
			}
			return arr
		}
		if v.IsObject() {
			obj := make(map[string]interface{})
			v.ForEach(func(key, val gjson.Result) bool {
				if cleaned := cleanValue(val); cleaned != nil {
					obj[key.String()] = cleaned
				}
				return true
			})
			if len(obj) == 0 {
				return nil
			}
			return obj
		}
	}
	return nil
}
