package constants

import "sync"

var (
	PARSE_SYNC_MAP_TO_MAP = func(s *sync.Map) map[string]interface{} {
		if s == nil {
			return nil
		}
		var m = make(map[string]interface{})
		s.Range(func(key, value any) bool {
			m[key.(string)] = value
			return true
		})
		return m
	}
)
