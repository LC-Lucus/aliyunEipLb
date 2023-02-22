package eip

import (
	"aliyunEipLb/aliyun"
	dbconfig "aliyunEipLb/config"
	"aliyunEipLb/log"
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Eip struct {
	gorm.Model
	State   int    `gorm:"type:int(4);default 0;comment:'0:正常 1:删除'"`
	Account string `gorm:"type:varchar(50);comment:'阿里云账号名称'"`
	// 必须添加unique才能用clause.OnConflict
	IpAddress       string `gorm:"type:varchar(30);unique;index:ip_addr;comment:'EIP的IP地址'"`
	Status          string `gorm:"type:varchar(10);comment:'EIP的状态:1、Associating：绑定中 2、Unassociating：解绑中 3、InUse：已分配 4、Available：可用 5、Releasing：释放中'"`
	InstanceId      string `gorm:"type:varchar(100);comment:'当前绑定的实例的ID'"`
	Name            string `gorm:"type:varchar(100);comment:'EIP的名称'"`
	RegionId        string `gorm:"type:varchar(50);comment:'EIP所在的地域ID'"`
	AllocationId    string `gorm:"type:varchar(100);comment:'EIP的实例ID'"`
	ResourceGroupId string `gorm:"type:varchar(100);comment:'资源组ID'"`
	Netmode         string `gorm:"type:varchar(30);comment:'网络类型。仅取值：public，表示公网'"`
	AllocationTime  string `gorm:"type:varchar(100);comment:'EIP的创建时间'"`
}

var IpAddresses []string
var Eips []Eip

// GetEip 获取弹性ip实例
func GetEip(a *aliyun.ALiYun) (err error) {
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(a.AK, a.SK)
	for _, r := range a.Region {
		client, err := vpc.NewClientWithOptions(r, config, credential)
		if err != nil {
			log.Errorf("创建客户端连接失败，原因：%v", err)
			return err
		}
		request := vpc.CreateDescribeEipAddressesRequest()

		request.Scheme = "https"
		response, err := client.DescribeEipAddresses(request)
		if err != nil {
			log.Errorf("查询弹性IP失败,原因：%v", err)
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
				r, err := client.DescribeEipAddresses(request)
				count += len(r.EipAddresses.EipAddress)
				if err != nil {
					log.Errorf("查询弹性IP失败,原因：%v", err)
					return err
				}
				for _, e := range r.EipAddresses.EipAddress {
					eip := Eip{
						Account:         a.Account,
						IpAddress:       e.IpAddress,
						Status:          e.Status,
						InstanceId:      e.InstanceId,
						Name:            e.Name,
						RegionId:        e.RegionId,
						AllocationId:    e.AllocationId,
						ResourceGroupId: e.ResourceGroupId,
						Netmode:         e.Netmode,
						AllocationTime:  e.AllocationTime,
					}
					//添加eip
					Eips = append(Eips, eip)
					//添加ipaddress
					IpAddresses = append(IpAddresses, eip.IpAddress)
				}
			}
		}
		if totalCount != count {
			log.Errorf("EIP数量查询有误，请重新查询！")
			errValue := fmt.Sprintf("EIP数量查询有误，请重新查询！")
			return errors.New(errValue)
		} else {
			InsertEips(Eips)
		}
	}
	return err
}

// InsertEips 向数据库中插入Eip信息
func InsertEips(eips []Eip) {
	// 创建表(优化：之前放到for循环中，每次都要进行判断表是否创建效率低下)
	dbconfig.DB.AutoMigrate(&Eip{})
	// 优化后，ip_address存在则更新，不存在则插入
	dbconfig.DB.Model(&Eip{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ip_address"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"state", "account", "status", "instance_id",
			"name", "region_id", "allocation_id",
			"resource_group_id", "netmode", "allocation_time", "updated_at",
		}),
	}).Create(&eips)

	// 优化前代码
	//for _, eip := range eips {
	//  //创建表(优化：之前放到for循环中，每次都要进行判断表是否创建效率低下)
	//  dbconfig.DB.AutoMigrate(&Eip{})
	//	//Find方法查询不到数据不会报错
	//	result := dbconfig.DB.Model(&Eip{}).Where("ip_address = ?", eip.IpAddress).First(&Eip{}).Error
	//	if result != nil {
	//		//不存在该数据则添加
	//		insertEips = append(insertEips, eip)
	//		dbconfig.DB.Create(&eip)
	//	} else {
	//		//存在该数据则修改
	//		updateEips = append(updateEips, eip)
	//		dbconfig.DB.Model(&Eip{}).Where("ip_address = ?", eip.IpAddress).Updates(eip)
	//	}
	//}
}

// IsDeletedEips 查询表内数据与新数据作对比新数据如果没有则标记为删除(软删除)
func IsDeletedEips(ipAddrs []string) {
	var eips []Eip
	dbconfig.DB.Select("ip_address").Find(&eips)
	for _, eip := range eips {
		flag := true
		for _, ip := range ipAddrs {
			if ip == eip.IpAddress {
				flag = false
				break
			}
		}
		if flag {
			dbconfig.DB.Model(&Eip{}).Where("ip_address = ?", eip.IpAddress).Update("state", 1)
			dbconfig.DB.Where("ip_address = ?", eip.IpAddress).Delete(&Eip{})
		}
	}
}

//获取每一条数据
//查询表内有没有该数据
//1、没有则插入 2、有则更新
//一条条查询表内数据
//与新数据作对比新数据如果没有则标记为删除
