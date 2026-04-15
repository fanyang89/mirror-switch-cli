package mirrors

type MirrorNode struct {
	Name string
	Host string
}

var MirrorNodes = []MirrorNode{
	{"CERNET (Redirection Service)", "mirrors.cernet.edu.cn"},
	{"TUNA (Tsinghua University)", "mirrors.tuna.tsinghua.edu.cn"},
	{"USTC (Univ. of Science and Technology of China)", "mirrors.ustc.edu.cn"},
	{"SJTUG (Shanghai Jiao Tong University)", "mirror.sjtu.edu.cn"},
	{"NJU (Nanjing University)", "mirrors.nju.edu.cn"},
	{"BFSU (Beijing Foreign Studies University)", "mirrors.bfsu.edu.cn"},
	{"JLU (Jilin University)", "mirrors.jlu.edu.cn"},
	{"SDU (Shandong University)", "mirrors.sdu.edu.cn"},
	{"LUG @ UESTC (Univ. of Electronic Science and Technology)", "mirrors.pku.edu.cn"}, // Note: PKU is also a major one
	{"SUSTech (Southern Univ. of Science and Technology)", "mirrors.sustech.edu.cn"},
	{"NWAFU (Northwest A&F University)", "mirrors.nwafu.edu.cn"},
	{"NJTech (Nanjing University of Technology)", "mirrors.njtech.edu.cn"},
	{"ZJU (Zhejiang University)", "mirrors.zju.edu.cn"},
}
