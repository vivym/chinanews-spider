package wordle

var regionMapR = map[string]string{
	"AH":   "安徽",
	"AM":   "澳门",
	"BJ":   "北京",
	"CQ":   "重庆",
	"FJ":   "福建",
	"GD":   "广东",
	"GS":   "甘肃",
	"GX":   "广西",
	"GZ":   "贵州",
	"HA":   "河南",
	"HB":   "湖北",
	"HEB":  "河北",
	"HI":   "海南",
	"HK":   "香港",
	"HLJ":  "黑龙江",
	"HN":   "湖南",
	"JL":   "吉林",
	"JS":   "江苏",
	"JX":   "江西",
	"LN":   "辽宁",
	"NMG":  "内蒙古",
	"NX":   "宁夏",
	"QH":   "青海",
	"SC":   "四川",
	"SD":   "山东",
	"SH":   "上海",
	"SHX":  "陕西",
	"SX":   "山西",
	"TW":   "台湾",
	"XJ":   "新疆",
	"XZ":   "西藏",
	"YN":   "云南",
	"ZJ":   "浙江",
	"R-CN": "中国",
	"R-DB": "东北",
	"R-HB": "华北",
	"R-HD": "华东",
	"R-HN": "华南",
	"R-HZ": "华中",
	"R-XB": "西北",
	"R-XN": "西南",
}

func regionToShort(region string) string {
	for key, value := range regionMapR {
		if value == region {
			return key
		}
	}
	return ""
}
