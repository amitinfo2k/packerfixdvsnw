# packerfixdvsnw

### Description : 

  Stop gap fix for VMWare ISO builder DVS switch

### Changes to be in Packer VMWare ISO Builder codebase :
 
File :  builder.go (src/github.com/mitchellh/packer/builder/vmware/iso/builder.go) 

* Add the following properties under Config struct

```sh
	FlagFixNetwork       bool     `mapstructure:"fix_vm_network"`
	VCenterSDKURL        string   `mapstructure:"vcenter_sdk_url"`
	Network              string   `mapstructure:"network"`
	DataCenter           string   `mapstructure:"datacenter"`
	NetworkType    	     string   `mapstructure:"network_type"`
    QualifiedVMName      string   `mapstructure:"ql_vm_name"`
```

* Call the StepFixNetwork after stepRegister and before stepRun.

```sh
	...
    &StepRegister{
			Format: b.config.Format,
	},
	&StepFixNetwork{
			FlagFixNetwork: b.config.FlagFixNetwork,
			VMName:      b.config.VMName,
			QualifiedVMName: b.config.QualifiedVMName,
			VCenterSDKURL: b.config.VCenterSDKURL,  
			Network:     b.config.Network,
			DataCenter:  b.config.DataCenter,
			NetworkType: b.config.NetworkType,	
	},
	&vmwcommon.StepRun{
	...
```

### Following are the properties need to be specified in packer template : 

```sh

  "fix_vm_network" : "true",
  "datacenter": "<datacenter-name>",
  "network_type": "vmxnet3|e1000",
  "network": "VLAN XXX",
  "vcenter_sdk_url": "https://user:passwd@x.x.x.x/sdk",

```
