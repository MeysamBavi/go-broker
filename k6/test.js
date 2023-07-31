import grpc from 'k6/net/grpc';
import { check } from 'k6';
import { b64encode } from 'k6/encoding';
import { Rate } from 'k6/metrics';

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
const failedConnects = new Rate(failed_connect_attempts);

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

const maxTarget = 5 * 7;
const batch = 100;
const timeUnit = '5s';

export const options = {
    vus: maxTarget,
    scenarios: {
        breakpoint: {
            executor: 'ramping-arrival-rate',
            startRate: 5 * 2,
            timeUnit: timeUnit,
            preAllocatedVUs: Math.floor(maxTarget * 105 / 100),
            stages: [
                {target: maxTarget, duration: '40s'},
                {target: maxTarget, duration: '5s'}
            ],
        },
    },
    thresholds: {
        grpc_req_duration: [
            {
                threshold: 'p(95) < 10000',
                abortOnFail: true,
                delayAbortEval: '5s'
            }
        ],
        failed_connect_attempts: [
            {
                threshold: 'rate <= 0.05',
                abortOnFail: true,
                delayAbortEval: '1s'
            }
        ],
        checks: [
            {
                threshold: 'rate >= 0.95',
                abortOnFail: true,
                delayAbortEval: '5s'
            }
        ],
        dropped_iterations: [
            {
                threshold: 'rate <= 5',
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
        client.connect('localhost:50043', {
            plaintext: true
        });
        failedConnects.add(false);
    } catch (err) {
        failedConnects.add(true);
        client.close();
        throw err;
    }

    for (let i = 0; i < batch; i++) {
        const subj = randomItem(subjects) + randomString(4, '01');
        const expirationSeconds = randomIntBetween(1, 20);
        const message = randomString(randomIntBetween(10, 60));

        const data = {subject: subj, body: b64encode(message), expirationSeconds: expirationSeconds};
        const response = client.invoke('broker.Broker/Publish', data);

        check(response, {
            'publish is OK': (r) => r && r.status === grpc.StatusOK,
        });

        const doFetch = randomIntBetween(1, 2) === 0;
        if (doFetch) {
            const data2 = {subject: subj, id: response.message.id};
            const response2 = client.invoke('broker.Broker/Fetch', data2);

            check(response2, {
                'fetch is OK': (r) => r && r.status === grpc.StatusOK,
            });
        }
    }

    client.close();
};
