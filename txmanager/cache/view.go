package cache

import (
	"github.com/Loopring/relay-lib/cache"
	"strings"
	"github.com/Loopring/relay-lib/log"
)

const TxViewSearchPreKey = "TXVIEW_"

func DelTxViewCacheByOwners(owners []string) {
	keyStrs := make([]string, 0)
	for _, owner := range owners {
		keys, _ := cache.Keys(strings.ToUpper(TxViewSearchPreKey + "OWNER:" + owner + "*"))
		for _, key := range keys {
			log.Debugf("[clear cache] clear key: %s", string(key))
			keyStrs = append(keyStrs, string(key))
		}
	}
	cache.Dels(keyStrs)
}
