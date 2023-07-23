import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { b64encode } from 'k6/encoding';
import { randomIntBetween, randomString, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

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

const client = new grpc.Client();
client.load(['../api/proto'], 'broker.proto');

export default () => {
    client.connect('localhost:50043', {
        plaintext: true
    });

    const subj = randomItem(subjects) + randomString(2, '01');
    const expirationSeconds = randomIntBetween(1, 20);
    const message = randomString(randomIntBetween(10, 60));

    const data = { subject: subj, body: b64encode(message), expirationSeconds: expirationSeconds };
    const response = client.invoke('broker.Broker/Publish', data);

    check(response, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    const doFetch = randomIntBetween(0, 2) === 0
    if (doFetch) {
        const data2 = { subject: subj, id: response.message.id };
        const response2 = client.invoke('broker.Broker/Fetch', data2);

        check(response2, {
            'status is OK': (r) => r && r.status === grpc.StatusOK,
        });
    }

    client.close()
};
