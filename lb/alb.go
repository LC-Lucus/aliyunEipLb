package lb

import (
	"aliyunEipLb/aliyun"
	dbconfig "aliyunEipLb/config"
	"aliyunEipLb/log"
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Alb struct {
	gorm.Model
	State              int    `gorm:"type:int(4);default 0;comment:'0:正常 1:删除'"`
	Account            string `gorm:"type:varchar(50);comment:'阿里云账号名称'"`
	LoadBalancerId     string `gorm:"type:varchar(100);unique;index:alb_id;comment:'负载均衡实例ID'"`
	LoadBalancerName   string `gorm:"type:varchar(100);comment:'负载均衡实例名称'"`
	LoadBalancerStatus string `gorm:"type:varchar(10);comment:'负载均衡的业务状态:Abnormal：异常 Normal：正常'"`
	ResourceGroupId    string `gorm:"type:varchar(100);comment:'企业资源组ID'"`
	DNSName            string `gorm:"type:varchar(100);comment:'DNS域名'"`
	VpcId              string `gorm:"type:varchar(100);comment:'负载均衡实例的专有网络ID'"`
	CreateTime         string `gorm:"type:varchar(100);comment:'资源创建时间'"`
}

var AlbLoadBalancerIds []string
var Albs []Alb

// GetALB 获取应用型负载均衡实例
func GetALB(a *aliyun.ALiYun) (err error) {
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(a.AK, a.SK)
	for _, r := range a.Region {
		client, err := alb.NewClientWithOptions(r, config, credential)
		if err != nil {
			log.Errorf("创建客户端连接失败，原因：%v", err)
			return err
		}
		request := alb.CreateListLoadBalancersRequest()

		request.Scheme = "https"
		response, err := client.ListLoadBalancers(request)
		if err != nil {
			log.Errorf("查询ALB失败,原因：%v", err)
			return err
		}

		// 分页
		count := 0
		totalCount := response.TotalCount
		if totalCount > 0 {
			for i := 0; i < totalCount/10+1; i++ {
				request.MaxResults = "10"
				r, err := client.ListLoadBalancers(request)
				count += len(r.LoadBalancers)
				request.NextToken = r.NextToken
				if err != nil {
					log.Errorf("查询ALB失败,原因：%v", err)
					return err
				}
				for _, l := range r.LoadBalancers {
					alb := Alb{
						Account:            a.Account,
						LoadBalancerId:     l.LoadBalancerId,
						LoadBalancerName:   l.LoadBalancerName,
						LoadBalancerStatus: l.LoadBalancerStatus,
						ResourceGroupId:    l.ResourceGroupId,
						DNSName:            l.DNSName,
						VpcId:              l.VpcId,
						CreateTime:         l.CreateTime,
					}
					//添加alb
					Albs = append(Albs, alb)
					//添加loadbalancerid
					AlbLoadBalancerIds = append(AlbLoadBalancerIds, alb.LoadBalancerId)
				}
			}
		}
		if totalCount != count {
			log.Errorf("ALB数量查询有误，请重新查询！")
			errValue := fmt.Sprintf("ALB数量查询有误，请重新查询！")
			return errors.New(errValue)
		} else {
			InsertAlbs(Albs)
		}
	}
	return err
}

// InsertAlbs 向数据库中插入Alb信息
func InsertAlbs(albs []Alb) {
	// 创建表
	dbconfig.DB.AutoMigrate(&Alb{})
	// 优化后，load_balancer_id存在则更新，不存在则插入
	dbconfig.DB.Model(&Alb{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "load_balancer_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"state", "account", "load_balancer_name",
			"load_balancer_status", "resource_group_id", "dns_name",
			"vpc_id", "create_time", "updated_at",
		}),
	}).Create(&albs)

}

// IsDeletedAlbs 查询表内数据与新数据作对比新数据如果没有则标记为删除(软删除)
func IsDeletedAlbs(albIds []string) {
	var albs []Alb
	dbconfig.DB.Select("load_balancer_id").Find(&albs)
	for _, alb := range albs {
		flag := true
		for _, albId := range albIds {
			if albId == alb.LoadBalancerId {
				flag = false
				break
			}
		}
		if flag {
			dbconfig.DB.Model(&Alb{}).Where("load_balancer_id = ?", alb.LoadBalancerId).Update("state", 1)
			dbconfig.DB.Where("load_balancer_id = ?", alb.LoadBalancerId).Delete(&Alb{})
		}
	}
}
