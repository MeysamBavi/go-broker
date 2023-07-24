import grpc from 'k6/net/grpc';
import { check } from 'k6';
import { b64encode } from 'k6/encoding';
import { Counter } from 'k6/metrics';

function randomIntBetween(min, max) { // min and max included
    return Math.floor(Math.random() * (max - min + 1) + min);
}

function randomItem(arrayOfItems){
    return arrayOfItems[Math.floor(Math.random() * arrayOfItems.length)];
}

function randomString(length, charset='abcdefghijklmnopqrstuvwxyz') {
    let res = '';
    while (length--) res += charset[(Math.random() * charset.length) | 0];
    return res;
}

const failed_connect_attempts = 'failed_connect_attempts';
const failedConnects = new Counter(failed_connect_attempts);

const subjects = [
    'alpha',
    'beta',
    'gamma',
    'delta',
    'epsilon',
    'zeta',
    'eta',
    'theta',
    'iota',
    'kappa',
];

export const options = {
    scenarios: {
        contacts: {
            executor: 'ramping-arrival-rate',

            // Start iterations per `timeUnit`
            startRate: 10,

            // Start `startRate` iterations per minute
            timeUnit: '1s',

            preAllocatedVUs: 50,

            maxVUs: 2000,

            stages: [
                {target: 10000, duration: '10m'},
            ],
        },
    },
    thresholds: {
        grpc_req_duration: [
            {
                threshold: 'p(95) < 10000',
                abortOnFail: true,
                delayAbortEval: '10s'
            }
        ],
        failed_connect_attempts: [
            {
                threshold: 'rate < 2',
                abortOnFail: true,
                delayAbortEval: '10s'
            }
        ]
    }
};

const client = new grpc.Client();
client.load(['../api/proto'], 'broker.proto');

export default () => {
    try {
        if (!client.isConnected) {
            client.connect('localhost:50043', {
                plaintext: true
            });
        }
    } catch (err) {
        failedConnects.add(1);
        throw err;
    }

    const subj = randomItem(subjects) + randomString(2, '01');
    const expirationSeconds = randomIntBetween(1, 20);
    const message = randomString(randomIntBetween(10, 60));

    const data = { subject: subj, body: b64encode(message), expirationSeconds: expirationSeconds };
    const response = client.invoke('broker.Broker/Publish', data);

    check(response, {
        'publish is OK': (r) => r && r.status === grpc.StatusOK,
    });

    const doFetch = randomIntBetween(0, 2) === 0
    if (doFetch) {
        const data2 = { subject: subj, id: response.message.id };
        const response2 = client.invoke('broker.Broker/Fetch', data2);

        check(response2, {
            'fetch is OK': (r) => r && r.status === grpc.StatusOK,
        });
    }

    client.close()
};
