package iso

import (
        "fmt"
        "reflect"
	"time"
        "golang.org/x/net/context"
	"net/url"
        "github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/types"
        "github.com/vmware/govmomi/object"
	"github.com/mitchellh/multistep"
        "github.com/mitchellh/packer/packer"
)

// This step runs the fixes the Network configuration for  virtual machine.
//
// Uses:
//   driver Driver
//   ui     packer.Ui
//   VCenterSDKURL string - https:/<username>:<password>@<vcenter-host>/sdk
//   Network string - network name/vlan name
//   DataCenter string - Data center name
//   NetworkType string - network type  (e1000 | vmxnet3)
//   VMName string - VN name
//   QualifiedVMName string - Qualified vmname can be specified in case govami not able to find the vm (optional)
//   FlagFixNetwork - bool - is this module need to enabled (optional)
// Produces:
//   <nothing>
type StepFixNetwork struct {
	
	VCenterSDKURL string
	Network string
	DataCenter string
	VMName string
	FlagFixNetwork bool
	NetworkType string
	QualifiedVMName string
}

func addVNIC(ui packer.Ui,f *find.Finder,ctx context.Context,c *govmomi.Client,vm *object.VirtualMachine,network string,nwType string ) error {
	
        ui.Say("Adding NIC")

	nets, err := f.NetworkList(ctx,network)
	if err != nil {
		return err
	}
        // TODO expose param for DVS
	net := nets[1]

	backing, err := net.EthernetCardBackingInfo(ctx)
	if err != nil {
		return err	
	}
	device, err := object.EthernetCardTypes().CreateEthernetCard(nwType, backing)
	if err != nil {
		return err 
	}
	err = vm.AddDevice(ctx, device)
	if err != nil {
		return err
	}
	ui.Say("Adding NIC Success")

return nil
}//

func delVNIC(ui packer.Ui,f *find.Finder,ctx context.Context,vm *object.VirtualMachine) error {
	ui.Say("Deleting NIC ")
	devicelst, err := vm.Device(ctx)
	if err != nil {
		return err
	}

	for _, device := range devicelst {

		switch device.(type) {
		case *types.VirtualVmxnet3:
			ui.Message(fmt.Sprintf("Removing NIC %s\n", device.GetVirtualDevice().DeviceInfo))
			err := vm.RemoveDevice(ctx, device)
			if err != nil {
			   return err
			}
			return nil

		case *types.VirtualE1000:
			ui.Message(fmt.Sprintf("Removing NIC %s\n", device.GetVirtualDevice().DeviceInfo))
			err := vm.RemoveDevice(ctx, device)
			if err != nil {
				return err
			}
			return nil
		default:
			fmt.Printf("Type %s\n", reflect.TypeOf(device).Elem())
			fmt.Printf("Device info %s\n", device.GetVirtualDevice().DeviceInfo)         
       
		}

	}
	
return nil
}//


func (s *StepFixNetwork) Run(state multistep.StateBag) multistep.StepAction {
        ui := state.Get("ui").(packer.Ui)
	ui.Say("Waiting for vm")
        time.Sleep(20 * time.Second)

        ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse URL from string
	// url.
	u, err := url.Parse(s.VCenterSDKURL)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// Connect and log in to ESX or vCenter
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	f := find.NewFinder(c.Client, true)

	ui.Say("Getting DataCenter ")

	// Find one and only datacenter
	dc, err := f.Datacenter(ctx, s.DataCenter)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("DataCenter Name : %s",dc.String()))

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)

	var qualifiedVMName = "/"+s.DataCenter+"/vm/Discovered virtual machine/"+s.VMName
   
    fmt.Printf("qualifiedVMName VM %s",qualifiedVMName)

	// Find vm in datacenter
	vm, err := f.VirtualMachine(ctx, qualifiedVMName)
	if err != nil {
		ui.Message(fmt.Sprintf("Retrying to get VM using QualifiedVMName VM %s",s.QualifiedVMName))
		vm, err = f.VirtualMachine(ctx, s.QualifiedVMName)
		
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	err = delVNIC(ui,f,ctx,vm)
	
        if err != nil {
          ui.Error(err.Error())
          return multistep.ActionHalt
        }
        ui.Say("Waiting for delete nic")
    	time.Sleep(20 * time.Second)

	err = addVNIC(ui,f,ctx,c,vm,s.Network,s.NetworkType)
	if err != nil {
          ui.Error(err.Error())
          return multistep.ActionHalt
        }
        ui.Say("Waiting for add nic")        
        time.Sleep(20 * time.Second)
		
return multistep.ActionContinue
}

func (s *StepFixNetwork) Cleanup(state multistep.StateBag) {
	// Nothing to be done for now
}



