# Full configuration options can be found at https://www.nomadproject.io/docs/configuration

datacenter = "run3"
data_dir = "/opt/nomad/data"
bind_addr = "0.0.0.0"

# This hosts name, just to avoid confusion with terraformed hosts
name = "run3"

# Increase log verbosity
#log_level = "DEBUG"
log_level = "INFO"
log_json = true
enable_syslog = true
#log_rotate_duration = "24h" !!! use this if putting logs somewhere other than syslog

server {
  # license_path is required as of Nomad v1.1.1+
  #license_path = "/etc/nomad.d/nomad.hcl"
  enabled = true
  bootstrap_expect = 1
}

advertise {
  http = "192.168.124.162:4646"
  rpc  = "192.168.124.162:4647"
  serf = "192.168.124.162:4648"
}

plugin "docker" {
  config {
    #endpoint = "unix:///var/run/docker.sock"
    volumes {
      enabled = true
    }
  }
}

client {
  enabled = true
  servers = ["127.0.0.1"]
  # 'node_class' used to ensure jobs meant for host 'run3' do run on 'run3'
  node_class = "run3"
  # run3 host has 8CPU's at 1.8GHz : so limit client to 4 CPU's worth:
  cpu_total_compute = 7200
  # run3 host has 8GB RAM : so limit client to 4GB:
  memory_total_mb = 4096

  host_volume "minio" {
    path = "/mnt/S3andSQS/tmp/minio/data"
    read_only = false
  }

  host_volume "localSQS" {
    # path = "/home/rhys/public/nomad-jobs/localSQS"
    path = "/mnt/S3andSQS/tmp/nomad/localSQS"
    read_only = false
  }
  host_volume "vaultUnsealTLS" {
    path = "/mnt/S3andSQS/tmp/nomad/vaultUnsealTLS"
    read_only = false
  }
}