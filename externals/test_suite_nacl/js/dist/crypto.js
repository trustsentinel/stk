"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tweetnacl_1 = require("tweetnacl");
const tweetnacl_util_1 = require("tweetnacl-util");
const fixedNonce = () => new Uint8Array([1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1]); //randomBytes(secretbox.nonceLength); 
const newNonce = () => tweetnacl_1.randomBytes(tweetnacl_1.secretbox.nonceLength);
exports.generateKey = () => tweetnacl_util_1.encodeBase64(tweetnacl_1.randomBytes(tweetnacl_1.secretbox.keyLength));
exports.encrypt = (keyUint8Array, msg) => {
    const nonce = newNonce();
    console.log("NONCE", nonce);
    const messageUint8 = tweetnacl_util_1.decodeUTF8(msg);
    const box = tweetnacl_1.secretbox(messageUint8, nonce, keyUint8Array);
    const fullMessage = new Uint8Array(nonce.length + box.length);
    fullMessage.set(nonce);
    fullMessage.set(box, nonce.length);
    console.log("fullMessage", fullMessage);
    const base64FullMessage = tweetnacl_util_1.encodeBase64(fullMessage);
    return base64FullMessage;
};
exports.decrypt = (keyUint8Array, messageWithNonce) => {
    const messageWithNonceAsUint8Array = tweetnacl_util_1.decodeBase64(messageWithNonce);
    const nonce = messageWithNonceAsUint8Array.slice(0, tweetnacl_1.secretbox.nonceLength);
    const message = messageWithNonceAsUint8Array.slice(tweetnacl_1.secretbox.nonceLength, messageWithNonce.length);
    const decrypted = tweetnacl_1.secretbox.open(message, nonce, keyUint8Array);
    if (!decrypted) {
        throw new Error("Could not decrypt message");
    }
    return tweetnacl_util_1.encodeUTF8(decrypted);
};
exports.buildSharedKey = (pub, priv) => tweetnacl_1.box.before(pub, priv);
/*
const obj = { hello: 'world' };
const pairA = generateKeyPair();
const pairB = generateKeyPair();
const sharedA = box.before(pairB.publicKey, pairA.secretKey);
const sharedB = box.before(pairA.publicKey, pairB.secretKey);
const encrypted = encrypt(sharedA, obj);
const decrypted = decrypt(sharedB, encrypted);
console.log(obj, encrypted, decrypted);*/ 
//# sourceMappingURL=crypto.js.map