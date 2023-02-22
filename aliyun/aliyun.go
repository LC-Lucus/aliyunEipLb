package aliyun

type ALiYun struct {
	AK      string   `json:"ak"`
	SK      string   `json:"sk"`
	Account string   `json:"account"`
	Region  []string `json:"region"`
}

func NewALiYun(ak, sk, account string, region []string) *ALiYun {
	return &ALiYun{
		AK:      ak,
		SK:      sk,
		Account: account,
		Region:  region,
	}
}
