function(opts)
  local pod_base = {
    kind: 'Pod',
    spec: {
      dnsPolicy: opts.DNSPolicy,
      containers: [
        {
          image: opts.ContainerImage,
          name: 'logger',
        },
      ],
    },
    apiVersion: 'v1',
    metadata: {
      name: 'logger-pod',
    },
  };
  [
    pod_base {
      metadata: {
        name: name,
      },
    }
    for name in ['pod1', 'pod2', 'pod3']
  ]
