package azure

type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	IsDefault bool   `json:"isDefault"`
}

type VM struct {
	Name          string            `json:"name"`
	ResourceGroup string            `json:"resourceGroup"`
	Location      string            `json:"location"`
	Tags          map[string]string `json:"tags"`
	HardwareProfile struct {
		VMSize string `json:"vmSize"`
	} `json:"hardwareProfile"`
	StorageProfile struct {
		ImageReference struct {
			Publisher string `json:"publisher"`
			Offer     string `json:"offer"`
			Sku       string `json:"sku"`
			Version   string `json:"version"`
		} `json:"imageReference"`
		OsDisk struct {
			OsType string `json:"osType"`
		} `json:"osDisk"`
	} `json:"storageProfile"`
	PublicIps  string   `json:"publicIps"`
	PrivateIps string   `json:"privateIps"`
	Fqdns      string   `json:"fqdns"`
	PowerState string   `json:"powerState"`
	Zones      []string `json:"zones"`
}

func (v *VM) OS() string {
	if v.StorageProfile.OsDisk.OsType != "" {
		return v.StorageProfile.OsDisk.OsType
	}
	return v.StorageProfile.ImageReference.Offer
}

func (v *VM) Image() string {
	ref := v.StorageProfile.ImageReference
	if ref.Publisher == "" {
		return ""
	}
	return ref.Publisher + ":" + ref.Offer + ":" + ref.Sku
}

func (v *VM) Zone() string {
	if len(v.Zones) > 0 {
		return v.Zones[0]
	}
	return ""
}
