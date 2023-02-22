package lb

import (
	"aliyunEipLb/aliyun"
	dbconfig "aliyunEipLb/config"
	"aliyunEipLb/log"
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Slb struct {
	gorm.Model
	State              int    `gorm:"type:int(4);default 0;comment:'0:正常 1:删除'"`
	Account            string `gorm:"type:varchar(50);comment:'阿里云账号名称'"`
	LoadBalancerId     string `gorm:"type:varchar(100);unique;index:slb_id;comment:'负载均衡实例ID'"`
	Address            string `gorm:"type:varchar(30);comment:'传统型负载均衡实例的服务地址'"`
	LoadBalancerName   string `gorm:"type:varchar(100);comment:'负载均衡实例的名称'"`
	LoadBalancerStatus string `gorm:"type:varchar(10);comment:'负载均衡实例状态:inactive：实例已停止 active：实例运行中 locked：实例已锁定'"`
	ResourceGroupId    string `gorm:"type:varchar(100);comment:'企业资源组ID'"`
	RegionId           string `gorm:"type:varchar(100);comment:'负载均衡实例的地域ID'"`
	NetworkType        string `gorm:"type:varchar(30);comment:'私网实例的网络类型: vpc：专有网络实例 classic：经典网络实例'"`
	VpcId              string `gorm:"type:varchar(100);comment:'私网负载均衡实例的专有网络ID'"`
	CreateTime         string `gorm:"type:varchar(100);comment:'实例创建时间'"`
}

var SlbLoadBalancerIds []string
var Slbs []Slb

// GetSLB 获取传统型负载均衡实例
func GetSLB(a *aliyun.ALiYun) (err error) {
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(a.AK, a.SK)
	for _, r := range a.Region {
		client, err := slb.NewClientWithOptions(r, config, credential)
		if err != nil {
			log.Errorf("创建客户端连接失败，原因：%v", err)
			return err
		}
		request := slb.CreateDescribeLoadBalancersRequest()

		request.Scheme = "https"
		response, err := client.DescribeLoadBalancers(request)
		if err != nil {
			log.Errorf("查询SLB失败,原因：%v", err)
			return err
		}
		// 分页
		count := 0
		totalCount := response.TotalCount
		if totalCount > 0 {
			for i := 0; i < totalCount/10+1; i++ {
				request.PageSize = "10"
				pageNumber := fmt.Sprintf("%d", i+1)
				request.PageNumber = requests.Integer(pageNumber) // 设定请求的PageNumber
				r, err := client.DescribeLoadBalancers(request)
				count += len(r.LoadBalancers.LoadBalancer)
				if err != nil {
					log.Errorf("查询SLB失败,原因：%v", err)
					return err
				}
				for _, l := range r.LoadBalancers.LoadBalancer {
					slb := Slb{
						Account:            a.Account,
						LoadBalancerId:     l.LoadBalancerId,
						Address:            l.Address,
						LoadBalancerName:   l.LoadBalancerName,
						LoadBalancerStatus: l.LoadBalancerStatus,
						ResourceGroupId:    l.ResourceGroupId,
						RegionId:           l.RegionId,
						NetworkType:        l.NetworkType,
						VpcId:              l.VpcId,
						CreateTime:         l.CreateTime,
					}
					//添加slb
					Slbs = append(Slbs, slb)
					//添加loadbalancerid
					SlbLoadBalancerIds = append(SlbLoadBalancerIds, slb.LoadBalancerId)
				}
			}
		}
		if totalCount != count {
			log.Errorf("SLB数量查询有误，请重新查询！")
			errValue := fmt.Sprintf("SLB数量查询有误，请重新查询！")
			return errors.New(errValue)
		} else {
			InsertSlbs(Slbs)
		}
	}
	return err
}

// InsertSlbs 向数据库中插入Slb信息
func InsertSlbs(slbs []Slb) {
	// 创建表
	dbconfig.DB.AutoMigrate(&Slb{})
	// 优化后，load_balancer_id存在则更新，不存在则插入
	dbconfig.DB.Model(&Slb{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "load_balancer_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"state", "account", "address", "load_balancer_name",
			"load_balancer_status", "resource_group_id", "region_id",
			"network_type", "vpc_id", "create_time", "updated_at",
		}),
	}).Create(&slbs)
}

// IsDeletedSlbs 查询表内数据与新数据作对比新数据如果没有则标记为删除(软删除)
func IsDeletedSlbs(slbIds []string) {
	var slbs []Slb
	dbconfig.DB.Select("load_balancer_id").Find(&slbs)
	for _, slb := range slbs {
		flag := true
		for _, slbId := range slbIds {
			if slbId == slb.LoadBalancerId {
				flag = false
				break
			}
		}
		if flag {
			dbconfig.DB.Model(&Slb{}).Where("load_balancer_id = ?", slb.LoadBalancerId).Update("state", 1)
			dbconfig.DB.Where("load_balancer_id = ?", slb.LoadBalancerId).Delete(&Slb{})
		}
	}
}
