function(opts)
  {
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
  }
