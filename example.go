package main

import "fmt"
import "github.com/liyinhgqw/typesafe-config/parse"

var input string = `
akka {
  loglevel = DEBUG
  loggers = ["akka.event.slf4j.Slf4jLogger"]
  event-handlers = ["akka.event.slf4j.Slf4jEventHandler"]

  # Log the complete configuration at INFO level when the actor system is started.
  # This is useful when you are uncertain of what configuration is used.
  log-config-on-start = on

  actor {
    provider = "akka.cluster.ClusterActorRefProvider"
    debug.receive = on
  }
  remote {
    netty.tcp {
      hostname = "127.0.0.1"
      port = 2554
      send-buffer-size = 30720000b
      receive-buffer-size = 30720000b
      maximum-frame-size = 10240000b
    }
    transport-failure-detector {
      # since TCP is used, increase heartbeat-interval generously
      # how often the node publishes heartbeats
      heartbeat-interval = 100 s   # default 4s
      # if after this duration no heartbeat has been received, you have a problem
      acceptable-heartbeat-pause = 250 s  # default 10s
    }
    retry-gate-closed-for = 2 s  # default 5s
  }
  cluster {
    seed-nodes = ["akka.tcp://ripak@127.0.0.1:2554"]
    roles = ["entity", "e2"]
    auto-down-unreachable-after = 600 s
    metrics.enabled = off
    failure-detector {
      heartbeat-interval = 100 s  # default 1s
      acceptable-heartbeat-pause = 250 s # default 3 s
      threshold = 8.0    # default 8.0
    }
    scheduler {
      # make it less than system's tick-duration to force start a new one
      tick-duration = 9 ms # default 33ms
      ticks-per-wheel = 512 # default 512
    }
    use-dispatcher = cluster-dispatcher
    jmx.enabled = on
  }
  contrib.cluster {
    sharding {
      role  = "entity"
      buffer-size = 500000000
    }
    pub-sub {
      gossip-interval = 3s  # default 1s
    }
  }
  test.single-expect-default = 10 s
}

cluster-dispatcher {
  type = "Dispatcher"
  executor = "fork-join-executor"
  fork-join-executor {
    parallelism-min = 2
    parallelism-max = 4
  }
}

akka.extensions = ["akka.contrib.pattern.DistributedPubSubExtension"]

ripak {
  env = "staging"
  persistence {
    enable = on
    purge = on
    snapshot.batch = 100
  }

  timeline.max = 8000

  follow.sync = off

  sdk-dispatcher = {
    # Dispatcher is the name of the event-based dispatcher
    type = Dispatcher
    # What kind of ExecutionService to use
    executor = "fork-join-executor"
    # Configuration for the fork join pool
    fork-join-executor {
      # Min number of threads to cap factor-based parallelism number to
      parallelism-min = 20
      # Parallelism (threads) ... ceil(available processors * factor)
      parallelism-factor = 2.0
      # Max number of threads to cap factor-based parallelism number to
      parallelism-max = 100
    }
  }

  web = {
    host = "0.0.0.0"
    port = 8080
  }
}


//akka.persistence.journal.leveldb.dir = "target/journal"
//akka.persistence.snapshot-store.local.dir = "target/snapshots"


akka.persistence.journal.plugin = "hbase-journal"
akka.persistence.snapshot-store.plugin = "hadoop-snapshot-store"
hbase-journal {

  # Partitions will be used to avoid the "hot write region" problem.
  # Set this to a number greater than the expected number of regions of your table.
  # WARNING: It is not supported to change the partition count when already written to a table (could miss some records)
  partition.count = 5

  # Name of the table to be used by the journal
  table = "rp:ripak_pubsub_messages_benchmark"

  # Name of the family to be used by the journal
  family = "message"

  # When performing scans, how many items to we want to obtain per one next(N) call.
  # This most notably affects the speed at which message replay progresses, fine tune this to match your cluster.
  scan-batch-size = 200

  # Dispatcher for fetching and replaying messages
  replay-dispatcher = "akka-hbase-persistence-replay-dispatcher"

  # Default dispatcher used by plugin
  plugin-dispatcher = "akka-hbase-persistence-dispatcher"
}

hadoop-snapshot-store {
  # select your preferred implementation based on your needs
  #
  # * HBase - snapshots stored together with snapshots; Snapshot size limited by Int.MaxValue bytes (currently)
  #
  # * HDFS *deprecated, will be separate project* - can handle HUGE snapshots;
  #          Can be easily dumped to local filesystem using Hadoop CL tools
  #          impl class is "akka.persistence.hbase.snapshot.HdfsSnapshotter"
  mode = "hbase"

  hbase {
    # Name of the table to be used by the journal
    table = "rp:ripak_pubsub_snapshots_benchmark"

    # Name of the family to be used by the journal
    family = "snapshot"
  }

  # Default dispatcher used by plugin
  plugin-dispatcher = "akka-hbase-persistence-dispatcher"
}

akka-hbase-persistence-dispatcher {
  type = Dispatcher
  executor = "thread-pool-executor"
  thread-pool-executor {
    core-pool-size-min = 15
    core-pool-size-factor = 2.0
    core-pool-size-max = 30
  }
  throughput = 200
}


akka-hbase-persistence-replay-dispatcher {
  type = Dispatcher
  executor = "thread-pool-executor"
  thread-pool-executor {
    core-pool-size-min = 15
    core-pool-size-factor = 2.5
    core-pool-size-max = 30
  }
  throughput = 200
}
`

func main() {
	result, err := parse.Parse("test", input)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(result.Root)
		conf := result.GetConfig()
		fmt.Println(conf.GetValue("akka-hbase-persistence-replay-dispatcher.thread-pool-executor.core-pool-size-min"))
		fmt.Println(conf.GetFloat("akka-hbase-persistence-replay-dispatcher.thread-pool-executor.core-pool-size-min"))
		fmt.Println(conf.GetBool("akka.actor.debug.receive"))
		fmt.Println(conf.GetString("akka-hbase-persistence-replay-dispatcher.type"))
		fmt.Println(conf.GetString("akka.cluster.auto-down-unreachable-after"))
		fmt.Println(conf.GetValue("akka.cluster.roles"))
		fmt.Println(conf.GetArray("akka.cluster.roles"))
		fmt.Println(conf.GetString("akka-hbase-persistence-replay-dispatcher.executor"))
		fmt.Println(conf.GetFloat("akka-hbase-persistence-replay-dispatcher.thread-pool-executor.core-pool-size-factor"))
	}
}
