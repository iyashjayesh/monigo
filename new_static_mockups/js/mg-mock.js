/* Preview harness — NOT for shipping.
 *
 * Stubs fetch with the repo's own sample payloads from static/API/Res/ so the
 * migrated pages can be reviewed without a running Go process. Values jitter
 * on each poll so the live behaviour, the sparkline fill and the staleness
 * escalation are all visible.
 *
 * Loaded only by the _preview-*.html files.
 */
(function () {
  'use strict';

  var t = 0;
  function jitter(base, amp, floor) {
    var v = base + (Math.sin(t / 3) + (Math.random() - 0.5)) * amp;
    return floor !== undefined && v < floor ? floor : v;
  }

  var STACKS = [
    'goroutine 1309 [running]:\ngithub.com/iyashjayesh/monigo/core.CollectGoRoutinesInfo()\n\t/vendor/github.com/iyashjayesh/monigo/core/profile.go:42 +0x40\ngithub.com/iyashjayesh/monigo/api.GetGoRoutinesStats({0x10525dd98, 0x140002ae620})\n\t/vendor/github.com/iyashjayesh/monigo/api/api.go:55 +0x28\nnet/http.HandlerFunc.ServeHTTP(0x1400019cb08)\n\t/usr/local/go/src/net/http/server.go:2136 +0x38\n',
    'goroutine 1 [IO wait, 11 minutes]:\ninternal/poll.runtime_pollWait(0x10c20cea0, 0x72)\n\t/usr/local/go/src/runtime/netpoll.go:343 +0xa0\nnet.(*TCPListener).Accept(0x14000140ca0)\n\t/usr/local/go/src/net/tcpsock.go:315 +0x2c\nnet/http.(*Server).Serve(0x140004bc000)\n\t/usr/local/go/src/net/http/server.go:3056 +0x2b8\nmain.main()\n\t/example-monigo/main.go:63 +0x184\n',
    'goroutine 43 [select, 11 minutes]:\ngithub.com/nakabonne/tstorage.NewStorage.func3()\n\t/vendor/github.com/nakabonne/tstorage/storage.go:256 +0xa4\ncreated by github.com/nakabonne/tstorage.NewStorage in goroutine 1\n\t/vendor/github.com/nakabonne/tstorage/storage.go:252 +0x764\n',
    'goroutine 44 [chan receive]:\ngithub.com/iyashjayesh/monigo/timeseries.SetDataPointsSyncFrequency.func2()\n\t/vendor/github.com/iyashjayesh/monigo/timeseries/mongiodb.go:136 +0x128\ncreated by github.com/iyashjayesh/monigo/timeseries.SetDataPointsSyncFrequency in goroutine 1\n\t/vendor/github.com/iyashjayesh/monigo/timeseries/mongiodb.go:132 +0x13c\n',
    'goroutine 45 [runnable]:\nnet.(*TCPListener).Accept(0x140000743a0)\n\t/usr/local/go/src/net/tcpsock.go:315 +0x2c\ngithub.com/iyashjayesh/monigo.StartDashboard(0x0)\n\t/vendor/github.com/iyashjayesh/monigo/monigo.go:147 +0x378\n',
    'goroutine 1310 [IO wait]:\nnet.(*netFD).Read(0x1400033e080)\n\t/usr/local/go/src/net/fd_posix.go:55 +0x28\nnet/http.(*connReader).backgroundRead(0x140001b4ff0)\n\t/usr/local/go/src/net/http/server.go:683 +0x40\ncreated by net/http.(*connReader).startBackgroundRead in goroutine 1309\n',
    'goroutine 1311 [chan receive]:\nmain.(*OrderWorker).drain(0x14000212a80)\n\t/example-monigo/worker/order.go:88 +0x1c4\ncreated by main.startWorkers in goroutine 1\n\t/example-monigo/worker/pool.go:41 +0xd8\n',
    'goroutine 1312 [chan receive]:\nmain.(*OrderWorker).drain(0x14000212b40)\n\t/example-monigo/worker/order.go:88 +0x1c4\ncreated by main.startWorkers in goroutine 1\n\t/example-monigo/worker/pool.go:41 +0xd8\n'
  ];

  function rec(name, value, unit) {
    return { record_name: name, record_value: value, record_unit: unit };
  }

  function metrics() {
    t++;
    var heapInuse = jitter(1.75, 0.22, 0.4);
    return {
      core_statistics: { goroutines: Math.round(jitter(8, 1.2, 4)), uptime: '43.03 m' },
      load_statistics: {
        service_cpu_load: jitter(0.04, 0.02, 0.005).toFixed(3) + '%',
        system_cpu_load: '11.21%',
        service_memory_load: '0.10%',
        system_memory_load: '78.61%'
      },
      cpu_statistics: { total_cores: 10, total_logical_cores: 10 },
      memory_statistics: {
        total_system_memory: '16.00 GB',
        gc_pause_duration: jitter(11.3, 2.4, 2).toFixed(2) + ' ms',
        gc_pause_duration_ms: jitter(11.3, 2.4, 2),
        stack_memory_usage: '704.00 KB',
        free_swap_memory: '1.25 GB',
        mem_stats_records: [
          rec('Alloc', 865.34375, 'KB'),
          rec('HeapSys', 15.3125, 'MB'),
          rec('HeapIdle', 13.5546875, 'MB'),
          rec('HeapInuse', heapInuse, 'MB'),
          rec('HeapReleased', 12.46875, 'MB'),
          rec('StackInuse', 704, 'KB'),
          rec('NumGC', 34, 'bytes')
        ],
        raw_mem_stats_records: []
      },
      heap_alloc_by_service: '865.34 KB',
      heap_alloc_by_system: '15.31 MB',
      network_io: { bytes_sent: 6356356846, bytes_received: 11929894738 },
      health: {
        system_health: {
          percent: 0, healthy: false,
          message: '[Oops] The Overall Health is in rough shape.',
          icon_msg: 'System usage exceeds allowed limits: CPU Usage 198.33% / 90.00%, Memory Usage 87.36% / 90.00%'
        },
        service_health: {
          percent: 97.53, healthy: true,
          message: '[Outstanding] The Overall Health is rocking it!',
          icon_msg: 'Service usage is within limits: CPU Usage 0.40% / 90.00%, Memory Usage 0.01% / 90.00%, Goroutines 7.00 / 100'
        }
      },
      goroutine_leak_report: {
        total_goroutines: 45, stale_goroutines: 2, growing_groups: 1,
        leak_suspected: true,
        message: '2 stale groups, 1 growing across 6 of 6 retained snapshots.',
        stale_threshold_minutes: 10, snapshots_retained: 6, snapshots_required: 6
      }
    };
  }

  var ROUTES = {
    '/service-info': function () {
      return {
        service_name: 'data-api',
        service_start_time: '2024-09-08T15:31:57.446856+05:30',
        go_version: 'go1.21.0',
        process_id: 94917,
        storage_type: 'disk',
        retention_period: '7d',
        storage_on_disk: '412 MB'
      };
    },
    '/metrics': metrics,
    '/go-routines-stats': function () {
      return {
        number_of_goroutines: Math.round(jitter(45, 2, 38)),
        stack_view: STACKS,
        leak_report: {
          total_goroutines: 45, stale_goroutines: 2, growing_groups: 1,
          leak_suspected: true,
          message: '2 stale groups, 1 growing across 6 of 6 retained snapshots.',
          stale_threshold_minutes: 10, snapshots_retained: 6, snapshots_required: 6,
          suspicious_groups: []
        }
      };
    },
    '/function': function () {
      return {
        heavyWork: {
          function_last_ran_at: new Date(Date.now() - 2000).toISOString(),
          memory_usage: 2244608, goroutine_count: 0, execution_time: 301000000
        },
        syncInventory: {
          function_last_ran_at: new Date(Date.now() - 18000).toISOString(),
          memory_usage: 19818086, goroutine_count: 4, execution_time: 1420000000
        },
        processOrder: {
          function_last_ran_at: new Date(Date.now() - 1000).toISOString(),
          memory_usage: 626688, goroutine_count: 1, execution_time: 42000000
        },
        validateInput: {
          function_last_ran_at: new Date(Date.now() - 500).toISOString(),
          memory_usage: 12288, goroutine_count: 0, execution_time: 800000
        },
        rebuildIndex: {
          function_last_ran_at: new Date(Date.now() - 180000).toISOString(),
          memory_usage: 43204608, goroutine_count: 2, execution_time: 884000000
        }
      };
    },
    '/function-details': function () {
      return {
        function_name: 'heavyWork',
        core_profile: {
          cpu_profile: 'Type: cpu\nDuration: 3.28s, Total samples = 3.24s (98.78%)\n' +
            'Showing nodes accounting for 3.20s, 98.77% of 3.24s total\n\n' +
            '      flat  flat%   sum%        cum   cum%\n' +
            '     2.88s 88.89% 88.89%      3.20s 98.77%  main.heavyWork\n' +
            '     0.32s  9.88% 98.77%      0.32s  9.88%  math.Sqrt',
          mem_profile: 'Type: alloc_space\nShowing nodes accounting for 18.4MB, 100% of 18.4MB total\n\n' +
            '      flat  flat%   sum%        cum   cum%\n' +
            '   18.40MB   100%   100%    18.40MB   100%  runtime.mallocgc'
        },
        function_code_trace: 'func heavyWork() {\n    var sum float64\n    for i := 0; i < 1e8; i++ {\n        sum += math.Sqrt(float64(i))\n    }\n}'
      };
    },
    '/service-metrics': function () {
      var out = [], base = Date.now() - 3600000;
      for (var i = 0; i < 60; i++) {
        out.push({
          time: new Date(base + i * 60000).toISOString(),
          value: {
            service_cpu_load: 0.04 + Math.sin(i / 5) * 0.02 + Math.random() * 0.01,
            HeapInuse: 1800000 + Math.sin(i / 7) * 300000 + Math.random() * 80000,
            goroutines: 45 + Math.round(Math.sin(i / 9) * 4)
          }
        });
      }
      return out;
    }
  };

  var real = window.fetch;
  window.fetch = function (url, options) {
    var path = String(url).split('?')[0].replace('/monigo/api/v1', '');
    var handler = ROUTES[path];
    if (!handler) { return real.apply(window, arguments); }
    return Promise.resolve({
      ok: true, status: 200,
      json: function () { return Promise.resolve(handler()); }
    });
  };

  console.info('[MoniGo] preview harness active — responses are stubbed.');
}());
