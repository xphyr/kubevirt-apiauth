module github.com/xphyr/listvms

go 1.15

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628

require (
	github.com/spf13/pflag v1.0.3
	k8s.io/apimachinery v0.20.2
	kubevirt.io/client-go v0.19.0
)
