##############################################################################
# Terraform Main IaC
##############################################################################

resource "tls_private_key" "vsi_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# DO NOT USE FOR PRODUCTION, NOT SECURE
resource "local_file" "vsi_pem" {
  content         = tls_private_key.vsi_key.private_key_pem
  filename        = "${path.module}/vsi.pem"
  file_permission = "0600"
}

resource "ibm_is_vpc" "vpc" {
  name = "${local.BASENAME}-vpc"
}

resource "ibm_is_security_group" "sg1" {
  name = "${local.BASENAME}-sg1"
  vpc  = ibm_is_vpc.vpc.id
}

# # allow all incoming network traffic on port 22
# resource "ibm_is_security_group_rule" "ingress_ssh_all" {
#   group     = ibm_is_security_group.sg1.id
#   direction = "inbound"
#   remote    = "0.0.0.0/0"

#   tcp {
#     port_min = 22
#     port_max = 22
#   }
# }

# allow ping
resource "ibm_is_security_group_rule" "ingress_icmp_all" {
  group     = ibm_is_security_group.sg1.id
  direction = "inbound"
  remote    = "0.0.0.0/0"

  icmp {}
}

# resource "ibm_is_security_group_rule" "egress_all" {
#   group     = ibm_is_security_group.sg1.id
#   direction = "outbound"
#   remote    = "0.0.0.0/0"
# }

resource "ibm_is_subnet" "subnet1" {
  name                     = "${local.BASENAME}-subnet1"
  vpc                      = ibm_is_vpc.vpc.id
  zone                     = local.ZONE
  total_ipv4_address_count = 256
}

data "ibm_is_image" "ubuntu" {
  # name = "ibm-ubuntu-20-04-minimal-amd64-2"
  name = "ibm-centos-8-2-minimal-amd64-2"
}

resource "ibm_is_ssh_key" "ssh_key_id" {
  name       = "${local.BASENAME}-sshkey"
  public_key = tls_private_key.vsi_key.public_key_openssh
}

resource "ibm_is_instance" "vsi1" {
  name    = "${local.BASENAME}-${var.vsi_name}"
  vpc     = ibm_is_vpc.vpc.id
  zone    = local.ZONE
  keys    = [ibm_is_ssh_key.ssh_key_id.id]
  image   = data.ibm_is_image.ubuntu.id
  profile = "cx2-2x4"

  primary_network_interface {
    name            = "foobar"
    subnet          = ibm_is_subnet.subnet1.id
    security_groups = [ibm_is_security_group.sg1.id]
  }

#   network_interfaces {
#     subnet          = ibm_is_subnet.subnet1.id
#     security_groups = [ibm_is_security_group.sg1.id]
#   }

#   network_interfaces {
#     subnet          = ibm_is_subnet.subnet1.id
#     security_groups = [ibm_is_security_group.sg1.id]
#   }
}

resource "ibm_is_floating_ip" "fip1" {
  name   = "${local.BASENAME}-fip1"
  target = ibm_is_instance.vsi1.primary_network_interface[0].id
}

# resource "null_resource" "post_exec" {
#   connection {
#     host        = ibm_is_floating_ip.fip1.address
#     private_key = tls_private_key.vsi_key.private_key_pem
#   }

#   provisioner "remote-exec" {
#     inline = [
#       "apt install vim"
#     ]
#   }
# }
