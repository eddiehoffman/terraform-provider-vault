package vault

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-vault/internal/provider"
)

const (
	KMSTypePKCS  = "pkcs11"
	KMSTypeAWS   = "awskms"
	KMSTypeAzure = "azurekeyvault"
)

func managedKeysResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: managedKeysWrite,
		DeleteContext: managedKeysDelete,
		ReadContext:   managedKeysRead,
		UpdateContext: managedKeysWrite,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allow_generate_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "If no existing key can be found in the referenced " +
					"backend, instructs Vault to generate a key within the backend",
			},

			"allow_store_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Controls the ability for Vault to import a key to the " +
					"configured backend, if 'false', those operations will be forbidden",
			},

			"any_mount": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Allow usage from any mount point within the namespace if 'true'",
			},
			"pkcs": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Configuration block for PKCS Managed Keys",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							Description: "A unique lowercase name that serves as " +
								"identifying the key",
						},
						"library": {
							Type:     schema.TypeString,
							Required: true,
							Description: "The name of the kms_library stanza to use from Vault's config " +
								"to lookup the local library path",
						},
						"key_label": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The label of the key to use",
						},
						"key_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The id of a PKCS#11 key to use",
						},
						"mechanism": {
							Type:     schema.TypeString,
							Required: true,
							Description: "The encryption/decryption mechanism to use, specified as a " +
								"hexadecimal (prefixed by 0x) string.",
						},
						"pin": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The PIN for login",
						},
						"slot": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "The slot number to use, specified as a string in a " +
								"decimal format (e.g. '2305843009213693953')",
						},
						"token_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The PIN for login",
						},
						"curve": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "Supplies the curve value when using " +
								"the 'CKM_ECDSA' mechanism. Required if " +
								"'allow_generate_key' is true",
						},
						"key_bits": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "Supplies the size in bits of the key when using " +
								"'CKM_RSA_PKCS_PSS', 'CKM_RSA_PKCS_OAEP' or 'CKM_RSA_PKCS' " +
								"as a value for 'mechanism'. Required if " +
								"'allow_generate_key' is true",
						},
						"force_rw_session": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The PIN for login",
						},
					},
				},
				MaxItems: 1,
			},
			"aws": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Configuration block for AWS Managed Keys",
				Elem: &schema.Resource{
					Schema: managedKeysAWSConfigSchema(),
				},
				MaxItems: 1,
			},
			"azure": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Configuration block for AWS Managed Keys",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							Description: "A unique lowercase name that serves as " +
								"identifying the key",
						},
						"tenant_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The tenant id for the Azure Active Directory organization",
						},
						"client_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The client id for credentials to query the Azure APIs",
						},
						"client_secret": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The client secret for credentials to query the Azure APIs",
						},
						"environment": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "AZUREPUBLICCLOUD",
							Description: "The Azure Cloud environment API endpoints to use",
						},
						"vault_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Key Vault vault to use the encryption keys for encryption and decryption",
						},
						"key_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The Key Vault key to use for encryption and decryption",
						},
						"resource": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "vault.azure.net",
							Description: "The Azure Key Vault resource's DNS Suffix to connect to",
						},
						"key_bits": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "The size in bits for an RSA key. This field is required " +
								"when 'key_type' is 'RSA' or when 'allow_generate_key' is true",
						},
						"key_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of key to use",
						},
					},
				},
				MaxItems: 1,
			},
		},
	}
}

func managedKeysAddCommonSchema(d *schema.ResourceData, data map[string]interface{}) map[string]interface{} {
	commonFields := []string{"allow_generate_key", "allow_store_key", "any_mount"}

	for _, k := range commonFields {
		if v, ok := d.GetOk(k); ok {
			data[k] = v.(string)
		}
	}

	return data
}

func managedKeysAWSConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			Description: "A unique lowercase name that serves as " +
				"identifying the key",
		},
		"access_key": {
			Type:     schema.TypeString,
			Required: true,
			Description: "The AWS access key to use. This can also " +
				"be provided with the 'AWS_ACCESS_KEY_ID' env variable",
		},
		"secret_key": {
			Type:     schema.TypeString,
			Required: true,
			Description: "The AWS secret key to use. This can also " +
				"be provided with the 'AWS_SECRET_ACCESS_KEY' env variable",
		},
		"curve": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "The curve to use for an ECDSA key. Used " +
				"when key_type is 'ECDSA'. Required if " +
				"'allow_generate_key' is true",
		},
		"endpoint": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Used to specify a custom AWS endpoint",
		},
		"key_bits": {
			Type:     schema.TypeString,
			Required: true,
			Description: "The size in bits for an RSA key. This " +
				"field is required when 'key_type' is 'RSA'",
		},
		"key_type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The type of key to use",
		},
		"kms_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "An identifier for the key",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "us-east-1",
			Description: "The AWS region where the keys are stored (or will be stored)",
		},
	}
}

func readAWSConfigBlock(d *schema.ResourceData) (string, map[string]interface{}) {
	data := map[string]interface{}{}

	blockField := "aws"
	for blockKey := range managedKeysAWSConfigSchema() {
		tfKey := fmt.Sprintf("%s.%d.%s", blockField, 0, blockKey)
		if v, ok := d.GetOk(tfKey); ok {
			data[blockKey] = v
		}
	}

	tfNameField := fmt.Sprintf("%s.%d.%s", blockField, 0, "name")
	name := d.Get(tfNameField).(string)

	return name, data
}

func getManagedKeysPath(keyType, name string) string {
	return fmt.Sprintf("sys/managed-keys/%s/%s", keyType, name)
}

func managedKeysWrite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, e := provider.GetClient(d, meta)
	if e != nil {
		return diag.FromErr(e)
	}

	if _, ok := d.GetOk("aws"); ok {
		awsKeyName, awsData := readAWSConfigBlock(d)
		awsKeyPath := getManagedKeysPath(KMSTypeAWS, awsKeyName)

		// add common schema fields
		awsData = managedKeysAddCommonSchema(d, awsData)

		if _, err := client.Logical().Write(awsKeyPath, awsData); err != nil {
			return diag.Errorf("error writing managed key %q, err=%s", awsKeyPath, err)
		}
	}

	// @TODO figure out what the ID should be
	// d.SetId(path)

	return managedKeysRead(ctx, d, meta)
}

func managedKeysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, e := provider.GetClient(d, meta)
	if e != nil {
		return diag.FromErr(e)
	}

	path := d.Id()

	resp, err := client.Logical().Read(path)
	if err != nil {
		return diag.FromErr(err)
	}

	fields := []string{"allow_generate_key", "allow_store_key", "any_mount"}

	for _, k := range fields {
		if v, ok := resp.Data[k]; ok {
			if err := d.Set(k, v); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func managedKeysDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, e := provider.GetClient(d, meta)
	if e != nil {
		return diag.FromErr(e)
	}
	path := d.Id()

	log.Printf("[DEBUG] Deleting managed key %s", path)
	_, err := client.Logical().Delete(path)
	if err != nil {
		return diag.Errorf("error deleting managed key %s", path)
	}
	log.Printf("[DEBUG] Deleted managed key %q", path)

	return nil
}
