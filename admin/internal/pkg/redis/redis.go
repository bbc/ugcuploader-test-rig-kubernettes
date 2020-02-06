package redis

import (
	"fmt"
	"strings"
	"time"

	types "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

//Redis used for all redis operations
type Redis struct {
	RedisPool *redis.Pool
}

//Setup used to setup the redis pool
func (r *Redis) Setup() {
	r.RedisPool = &redis.Pool{
		MaxIdle:     100,               // adjust to your needs
		IdleTimeout: 240 * time.Second, // adjust to your needs
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}

}

//AddTenant used to set the tenant started state to not-yet
func (r *Redis) AddTenant(redisTenant types.RedisTenant) (error string, started bool) {
	return r.addItem(fmt.Sprintf("testStartStatus:%s", redisTenant.Tenant), redisTenant)
}

//AddToWaitingForDelete used to add the tenant to list of waiting to be deleted
func (r *Redis) AddToWaitingForDelete(redisTenant types.RedisTenant) (error string, added bool) {
	return r.addItem(fmt.Sprintf("testDeletionStatus:%s", redisTenant.Tenant), redisTenant)
}

func (r *Redis) addItem(item string, redisTenant types.RedisTenant) (error string, started bool) {
	c := r.RedisPool.Get()
	defer c.Close()
	if _, err := c.Do("HMSET", redis.Args{}.Add(item).AddFlat(&redisTenant)...); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems adding tenant to redis")
		error = err.Error()
		started = false
		return
	}
	r.addTTL(item, 3600)
	started = true
	return

}

//GetTenantFromDelete used to get the tenant from the hash of deletes
func (r *Redis) GetTenantFromDelete(tenant string) (ten types.RedisTenant, error string, found bool) {
	return r.getItem(fmt.Sprintf("testDeletionStatus:%s", tenant))
}

//GetTenant used to get the tenant
func (r *Redis) GetTenant(tenant string) (ten types.RedisTenant, error string, found bool) {
	return r.getItem(fmt.Sprintf("testStartStatus:%s", tenant))
}

func (r *Redis) getItem(item string) (ten types.RedisTenant, error string, found bool) {
	c := r.RedisPool.Get()
	redisTenant := types.RedisTenant{}
	values, err := redis.Values(c.Do("HGETALL", item))
	defer c.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems getting tenant from redis")
		error = err.Error()
		found = false
	}
	redis.ScanStruct(values, &redisTenant)
	ten = redisTenant
	found = true
	return
}

//RemoveTenantDelete remove tenant from delete list
func (r *Redis) RemoveTenantDelete(tenant string) (error string, deleted bool) {
	return r.removeItem(fmt.Sprintf("testDeletionStatus:%s", tenant))
}

//RemoveTenant used to remove the tenant test status
func (r *Redis) RemoveTenant(tenant string) (error string, deleted bool) {
	return r.removeItem(fmt.Sprintf("testStartStatus:%s", tenant))
}

func (r *Redis) removeItem(item string) (error string, deleted bool) {
	c := r.RedisPool.Get()
	defer c.Close()
	_, err := c.Do("DEL", item)
	defer c.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems getting tenant from redis")
		error = err.Error()
		deleted = false
	}

	deleted = true
	return
}

//addTTL used to add and expiration time to the key
func (r *Redis) addTTL(key string, ttl int) (error string, added bool) {
	c := r.RedisPool.Get()
	defer c.Close()
	_, err := redis.Int(c.Do("EXPIRE", key, ttl))
	if err != nil {
		error = err.Error()
		added = false
	}

	added = true
	return
}

//AddToListOfStarted add to tenant to list of started test
func (r *Redis) AddToListOfStarted(tenant string) (error string, added bool) {
	return r.addItemToList("waitingForStartToComplete", tenant)
}

//BeingDeleted this adds tenants that are being deleted
func (r *Redis) BeingDeleted(tenant string) (error string, added bool) {
	return r.addItemToList("waitingForDeleteToComplete", tenant)
}

func (r *Redis) addItemToList(list string, item string) (error string, added bool) {
	c := r.RedisPool.Get()
	defer c.Close()
	if _, err := c.Do("SADD", list, item); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems adding tenant to deletion list to redis")
		error = err.Error()
		added = false
		return
	}
	r.addTTL(list, 3600)
	added = true
	return

}

//CheckIfBeingDeleted this adds tenants that are being deleted
func (r *Redis) CheckIfBeingDeleted(tenant string) (beingDeleted bool, error string, probs bool) {
	return r.checkIfItemExistInList("waitingForDeleteToComplete", tenant)
}

func (r *Redis) checkIfItemExistInList(list string, item string) (exist bool, error string, probs bool) {
	c := r.RedisPool.Get()
	defer c.Close()
	exist = false
	tenants, err := redis.Strings(c.Do("SMEMBERS", list))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems fetching list of tests waiting to be deleted")
		error = err.Error()
		probs = true
	}

	fmt.Printf(fmt.Sprintf("------------------ tenants in delete:%s: item %s", strings.Join(tenants, ","), item))

	for _, t := range tenants {
		if strings.EqualFold(t, item) {
			exist = true
			probs = false
			return
		}
	}
	probs = false
	return
}

//FetchWaitingToBeDeleted fetches test waiting to be deleted
func (r *Redis) FetchWaitingToBeDeleted() (tests []string, error string, found bool) {
	return r.getAllItemsInList("waitingForDeleteToComplete")
}

//FetchWaitingTests used to fetch the list of test waiting to start
func (r *Redis) FetchWaitingTests() (tests []string, error string, found bool) {
	return r.getAllItemsInList("waitingForStartToComplete")
}

func (r *Redis) getAllItemsInList(list string) (tests []string, error string, found bool) {
	c := r.RedisPool.Get()
	defer c.Close()
	values, err := redis.Strings(c.Do("SMEMBERS", list))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems fetching list of tests waiting to start")
		error = err.Error()
		found = false
	}
	tests = values
	found = true
	return
}

//RemoveFromWaitingTests used to remove test from waitting
func (r *Redis) RemoveFromWaitingTests(tenant string) (error string, removed bool) {
	return r.removeFromList("waitingForStartToComplete", tenant)
}

//RemoveFromWaitingForDelete used to remove from items waiting to be delete
func (r *Redis) RemoveFromWaitingForDelete(tennat string) (error string, removed bool) {
	return r.removeFromList("waitingForDeleteToComplete", tennat)
}

func (r *Redis) removeFromList(list string, item string) (error string, removed bool) {
	c := r.RedisPool.Get()
	defer c.Close()
	if _, err := c.Do("SREM", list, item); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems removing test from list")
		error = err.Error()
		removed = false
	}
	removed = true
	return

}
