package cistore

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	cloudinfo "github.com/banzaicloud/cloudinfo/internal/cloudinfo"
	"github.com/banzaicloud/cloudinfo/internal/cloudinfo/types"
	"github.com/banzaicloud/cloudinfo/internal/platform/modernredis"
	"github.com/go-redis/redis"
)

type RedisCache struct {
	client *redis.Client
	log    cloudinfo.Logger
}

func NewRedisCache(config modernredis.Config, log cloudinfo.Logger) cloudinfo.CloudInfoStore {
	client := modernredis.NewRedisClient(config)
	return &RedisCache{
		client: client,
		log:    log.WithFields(map[string]interface{}{"cistore": "modernredis"}),
	}
}

func (c *RedisCache) Ready() bool {
	return true
}

func (c *RedisCache) getKey(keyTemplate string, args ...interface{}) string {
	key := fmt.Sprintf(keyTemplate, args...)

	return key
}

func (c *RedisCache) DeleteRegions(provider, service string) {
	c.delete(c.getKey(cloudinfo.RegionKeyTemplate, provider, service))
}

func (c *RedisCache) DeleteZones(provider, service, region string) {
	c.delete(c.getKey(cloudinfo.ZoneKeyTemplate, provider, service, region))
}

func (c *RedisCache) DeleteImage(provider, service, regionId string) {
	c.delete(c.getKey(cloudinfo.ImageKeyTemplate, provider, service, regionId))
}

func (c *RedisCache) DeleteVersion(provider, service, region string) {
	c.delete(c.getKey(cloudinfo.VersionKeyTemplate, provider, service, region))
}

func (c *RedisCache) DeleteVm(provider, service, region string) {
	c.delete(c.getKey(cloudinfo.VmKeyTemplate, provider, service, region))
}

func (c *RedisCache) Export(w io.Writer) error {
	return nil
}

func (c *RedisCache) Import(r io.Reader) error {
	return nil
}

func (c *RedisCache) StoreRegions(provider, service string, val map[string]string) {
	c.put(c.getKey(cloudinfo.RegionKeyTemplate, provider, service), val)
}

func (c *RedisCache) GetRegions(provider, service string) (map[string]string, bool) {
	var (
		res = make(map[string]string)
	)
	ok, _ := c.get(c.getKey(cloudinfo.RegionKeyTemplate, provider, service), &res)

	return res, ok
}

func (c *RedisCache) StoreZones(provider, service, region string, val []string) {
	c.put(c.getKey(cloudinfo.ZoneKeyTemplate, provider, service, region), val)
}

func (c *RedisCache) GetZones(provider, service, region string) ([]string, bool) {
	var (
		res = make([]string, 0)
	)

	ok, _ := c.get(c.getKey(cloudinfo.ZoneKeyTemplate, provider, service, region), &res)
	return res, ok
}

func (c *RedisCache) StorePrice(provider, region, instanceType string, val types.Price) {
	c.put(c.getKey(cloudinfo.PriceKeyTemplate, provider, region, instanceType), val)
}

func (c *RedisCache) GetPrice(provider, region, instanceType string) (types.Price, bool) {
	var (
		res = types.Price{}
	)
	ok, _ := c.get(c.getKey(cloudinfo.PriceKeyTemplate, provider, region, instanceType), &res)

	return res, ok
}

func (c *RedisCache) StoreVm(provider, service, region string, val []types.VMInfo) {
	c.put(c.getKey(cloudinfo.VmKeyTemplate, provider, service, region), val)
}

func (c *RedisCache) GetVm(provider, service, region string) ([]types.VMInfo, bool) {
	var (
		res = make([]types.VMInfo, 0)
	)
	ok, _ := c.get(c.getKey(cloudinfo.VmKeyTemplate, provider, service, region), &res)

	return res, ok
}

func (c *RedisCache) StoreImage(provider, service, regionId string, val []types.Image) {
	c.put(c.getKey(cloudinfo.ImageKeyTemplate, provider, service, regionId), val)
}

func (c *RedisCache) GetImage(provider, service, regionId string) ([]types.Image, bool) {
	var (
		res = make([]types.Image, 0)
	)
	ok, _ := c.get(c.getKey(cloudinfo.ImageKeyTemplate, provider, service, regionId), &res)

	return res, ok
}

func (c *RedisCache) StoreVersion(provider, service, region string, val []types.LocationVersion) {
	c.put(c.getKey(cloudinfo.VersionKeyTemplate, provider, service, region), val)
}

func (c *RedisCache) GetVersion(provider, service, region string) ([]types.LocationVersion, bool) {
	var (
		res = make([]types.LocationVersion, 0)
	)
	ok, _ := c.get(c.getKey(cloudinfo.VersionKeyTemplate, provider, service, region), &res)

	return res, ok
}

func (c *RedisCache) StoreStatus(provider string, val string) {
	c.put(c.getKey(cloudinfo.StatusKeyTemplate, provider), val)
}

func (c *RedisCache) GetStatus(provider string) (string, bool) {
	var (
		res string
	)
	ok, _ := c.get(c.getKey(cloudinfo.StatusKeyTemplate, provider), &res)

	return res, ok
}

func (c *RedisCache) StoreServices(provider string, services []types.Service) {
	c.put(c.getKey(cloudinfo.ServicesKeyTemplate, provider), services)
}

func (c *RedisCache) GetServices(provider string) ([]types.Service, bool) {
	var (
		res = make([]types.Service, 0)
	)
	ok, _ := c.get(c.getKey(cloudinfo.ServicesKeyTemplate, provider), &res)

	return res, ok
}

func (c *RedisCache) Close() {
}

func (c *RedisCache) put(key string, value interface{}) error {
	return c.putEx(key, 0, value)
}

func (c *RedisCache) putEx(key string, expiration time.Duration, value interface{}) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = c.client.Set(key, val, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisCache) get(key string, out interface{}) (bool, error) {
	val, err := c.client.Get(key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if out != nil {
		err = json.Unmarshal([]byte(val), out)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (c *RedisCache) delete(key string) error {
	_, err := c.client.Del(key).Result()
	return err
}
