job "vault-unseal-tls" {
  # 'datacenters' is for human convenience and has no binding significance
  datacenters = ["run3"]
  type = "service"

  # All tasks in this job must run on client agent for host 'run3'
  constraint {
    attribute = "${node.class}"
    value     = "run3"
  }

  # The 'Task Group' name has the name of host 'run3' to indicate the host
  # (or the type of host) this job is on:
  group "run3" {
    count = 1

    # reduce default disk from 300MB as this does not log much
    ephemeral_disk {
      size = 200
      sticky = true
      migrate = false
    }
    

    # setup access to directory on host that has a 'host_volume' section
    # in '/etc/nomad.d/nomad.hcl' of:
    #     path = "/mnt/S3andSQS/tmp/nomad/vaultUnsealTLS"
    volume "vaultUnsealTLS" {
      type = "host"
      read_only = false
      source = "vaultUnsealTLS"
    }

    task "vault-unseal-tls" {
      driver = "docker"

      volume_mount {
        volume = "vaultUnsealTLS"
        # connect host directory with in container directory '/v2'
        destination = "/v2"
        read_only = false
      }

      config {
        image      = "alpine:3.20.0"
        force_pull = false

        ulimit {
          # ensure all memory can be locked, typically of use with java apps like elastic search
          memlock = "-1"
          # ensure enough open file handles can be created
          nofile = "65536"
          # ensure enough threads can be created
          #nproc = "65536"
          nproc = "8192"
        }

        command      = "sh"
        # The following if for home lab testing and NOT for production as its using unseal keys and certs.
        # The home lab servers are not left running all the time and vault needs unsealing after power up.
        # NOTE: the following starts with a 12 second delay to ensure nomad sees this task
        #       as alive for >=min_healthy_time
        #       otherwise it might think the task is dead and not run it as desired.
        args         = ["-c", "sleep 14;apk --update --no-cache add bash;ls v2;cd v2;./vault-unsealer-TLS;"]
        network_mode = "host"
      }

      resources {
        cpu    = 100
        memory = 50
      }
    }
  }
}
