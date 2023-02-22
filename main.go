package main

import (
	"aliyunEipLb/aliyun"
	"aliyunEipLb/config"
	"aliyunEipLb/eip"
	"aliyunEipLb/lb"
)

func main() {
	config.Setup("config/settings.yml")
	for _, i := range config.AliyunConfigs {
		AK := i.AK
		SK := i.SK
		Account := i.Account
		Region := i.Region
		aLiYunClient := aliyun.NewALiYun(AK, SK, Account, Region)
		//err := lb.GetSLB(aLiYunClient)
		//if err != nil {
		//	errValue := fmt.Sprintf("获取SLB失败，%v", err)
		//	log.Error(errValue)
		//	panic(errValue)
		//}
		//err = lb.GetALB(aLiYunClient)
		//if err != nil {
		//	errValue := fmt.Sprintf("获取ALB失败，%v", err)
		//	log.Error(errValue)
		//	panic(errValue)
		//}
		// 收集eip信息并做入库处理
		//lb.GetSLB(aLiYunClient)
		//eip.GetEip(aLiYunClient)

		//lb.GetALB(aLiYunClient)
		eip.GetEip(aLiYunClient)
		// 收集slb信息并做入库处理
		lb.GetSLB(aLiYunClient)
		// 收集alb信息并做入库处理
		lb.GetALB(aLiYunClient)
	}
	//判断阿里云上的eip是否被删除，如果被删除则在数据库中进行标记
	eip.IsDeletedEips(eip.IpAddresses)
	//判断阿里云上的slb是否被删除，如果被删除则在数据库中进行标记
	lb.IsDeletedSlbs(lb.SlbLoadBalancerIds)
	//判断阿里云上的alb是否被删除，如果被删除则在数据库中进行标记
	lb.IsDeletedAlbs(lb.AlbLoadBalancerIds)

}
