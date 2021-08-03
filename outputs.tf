##############################################################################
# Terraform Outputs
##############################################################################

output "sshcommand" {
  value = "ssh -i vsi.pem root@${ibm_is_floating_ip.fip1.address}"
}

output "public_ip" {
  value = ibm_is_floating_ip.fip1.address
}