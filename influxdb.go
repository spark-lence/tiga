package tiga

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxdbDao struct {
	client influxdb2.Client
	bucket string
	org    string
	config Configuration
}

func NewInfluxdbDao(config Configuration) *InfluxdbDao {
	env := config.GetEnv()
	host := config.GetConfigByEnv(env, "influxdb.host")
	port := config.GetConfigByEnv(env, "influxdb.port")
	bucket := config.GetConfigByEnv(env, "influxdb.bucket").(string)
	token := config.GetConfigByEnv(env, "influxdb.token").(string)
	org := config.GetConfigByEnv(env, "influxdb.org").(string)
	client := influxdb2.NewClientWithOptions(fmt.Sprintf("http://%s:%d", host, port), token, influxdb2.DefaultOptions().SetUseGZip(true).SetMaxRetries(3))
	return &InfluxdbDao{
		client: client,
		bucket: bucket,
		org:    org,
		config: config,
	}
}

func (i *InfluxdbDao) Write(measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error {
	writeAPI := i.client.WriteAPIBlocking(i.org, i.bucket)
	point := write.NewPoint(measurement, tags, fields, timestamp)
	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
		return fmt.Errorf("Data tags:%v,fields:%v wirte to %s,error:%w", tags, fields, measurement, err)
	}
	return nil

}
