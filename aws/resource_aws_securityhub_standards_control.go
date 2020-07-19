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
			"standards_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"disabled_reason": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsSecurityHubStandardsControlCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Enabling Security Hub standard control %s", d.Get("standards_control_arn"))

	var isEnabled string
	if enabled, ok := d.GetOk("enabled"); ok && enabled.(bool) {
		isEnabled = "ENABLED"
	} else {
		isEnabled = "DISABLED"
	}

	resp, err := conn.UpdateStandardsControl(&securityhub.UpdateStandardsControlInput{
		ControlStatus:       aws.String(isEnabled),
		DisabledReason:      aws.String(d.Get("disabled_reason").(string)),
		StandardsControlArn: aws.String(d.Get("standards_control_arn").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error setting Security Hub standard control to %s: %s", isEnabled, err)
	}

	log.Printf("[DEBUG] Response body: %s", resp)

	d.SetId(d.Get("standards_control_arn").(string))

	return resourceAwsSecurityHubStandardsControlRead(d, meta)
}

func resourceAwsSecurityHubStandardsControlRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Reading Security Hub standard control %s", d.Id())
	// https://gist.github.com/eferro/651fbb72851fa7987fc642c8f39638eb
	// will need to filter here, can't get a control directly. Need to record which standards the control is a part of too.
	resp, err := conn.DescribeStandardsControls(&securityhub.DescribeStandardsControlsInput{
		StandardsSubscriptionArn: aws.String(d.Get("standards_arn").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error reading control %s for Security Hub standard %s: %s", d.Id(), d.Get("standards_arn").(string), err)
	}

	if len(resp.Controls) != 1 {
		log.Printf("[WARN] Security Hub standard control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("enabled", resp.Controls[0].ControlStatus) // ENABLED/DISABLED -> true/false
	// if disabled, set disabled_reason.

	return nil
}

func resourceAwsSecurityHubStandardsControlDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Enabling Security Hub standard control %s", d.Id())

	_, err := conn.UpdateStandardsControl(&securityhub.UpdateStandardsControlInput{
		ControlStatus:       aws.String("ENABLED"),
		StandardsControlArn: aws.String(d.Get("standards_control_arn").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error disabling Security Hub standard control %s: %s", d.Id(), err)
	}

	return nil
}
