import { encrypt, decrypt, buildSharedKey} from './crypto'

export const A = {
    publicKey: new Uint8Array([16,44,122,53,10,77,214,164,220,19,151,64,250,13,191,5,61,6,88,138,89,1,190,18,119,176,132,39,227,233,73,35]),
    secretKey: new Uint8Array([160,184,150,49,86,171,26,150,235,114,62,19,126,123,44,11,225,238,103,124,241,203,122,13,147,247,18,48,24,232,121,121])
}

export const B = {
    publicKey: new Uint8Array([149,159,254,46,1,184,215,240,224,123,198,64,123,80,0,135,113,236,213,76,117,201,253,66,12,214,30,129,219,32,11,87]),
    secretKey: new Uint8Array([19,30,255,99,200,249,241,57,246,28,248,18,143,230,19,102,60,194,104,10,96,5,33,37,241,157,163,58,223,97,192,204])
}

const S = buildSharedKey(A.publicKey, B.secretKey)
console.log("Shared key, \nA:", A.publicKey, "\nB:", B.publicKey, "\na:", A.secretKey, "\nb:", B.secretKey, "\nS:", S)