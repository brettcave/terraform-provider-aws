package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSecurityHubStandardsControl() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubStandardsControlCreate,
		Read:   resourceAwsSecurityHubStandardsControlRead,
		Delete: resourceAwsSecurityHubStandardsControlDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"standards_control_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"enabled": {
				Type:		  schema.TypeBool,
				Optional:	  true,
				Default: 	  true,
			},
			"disabled_reason": {
				Type:		schema.TypeString,
				optional:	true,
			}
		},
	}
}

func resourceAwsSecurityHubStandardsControlCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Enabling Security Hub standard control %s", d.Get("standards_control_arn"))

	if d.GetOk("enabled") {
		isEnabled = "ENABLED"
	} else {
		isEnabled = "DISABLED"
	}

	resp, err := conn.UpdateStandardsControl(&securityhub.UpdateStandardsControlsInput{
		ControlStatus: isEnabled,
		DisabledReason: d.Get("disabled_reason"),
		StandardsControlArn: d.Get("standards_control_arn"),
	})

	if err != nil {
		return fmt.Errorf("Error setting Security Hub standard control to %s: %s", isEnabled, err)
	}

	standardsControl := resp.StandardsControls[0]

	d.SetId(*standardsControl.StandardsControlArn)

	return resourceAwsSecurityHubStandardsControlRead(d, meta)
}

func resourceAwsSecurityHubStandardsControlRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Reading Security Hub standard control %s", d.Id())
	// https://gist.github.com/eferro/651fbb72851fa7987fc642c8f39638eb
	// will need to filter here, can't get a control directly. Need to record which standards the control is a part of too.
	resp, err := conn.DescribeStandardsControls(&securityhub.DescribeStandardsControlsInput{
		StandardsControlArn: []*string{aws.String(d.Id())},
	})

	if err != nil {
		return fmt.Errorf("Error reading Security Hub standard control %s: %s", d.Id(), err)
	}

	if len(resp.StandardsControls) == 0 {
		log.Printf("[WARN] Security Hub standard control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	standardsControl := resp.StandardsControls[0]

	d.Set("standards_control_arn", standardsControl.StandardsArn)

	return nil
}

func resourceAwsSecurityHubStandardsControlDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Enabling Security Hub standard control %s", d.Id())

	_, err := conn.UpdateStandardsControl(&securityhub.UpdateStandardsControlInput{
		ControlStatus: "ENABLED",
		StandardsControlArn: d.Get("standards_control_arn"),
	})

	if err != nil {
		return fmt.Errorf("Error disabling Security Hub standard control %s: %s", d.Id(), err)
	}

	return nil
}
