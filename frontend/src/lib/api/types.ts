export interface UserRegistration {
    username: string;
    password: string;
};

export interface AuthResponse {
    userID: string;
};

export interface FieldError {
    field: string;
    code: 'alreadyExists' | 'minLength' | 'maxLength' | 'required' | 'invalidFormat' | 'unknown';
    message: string;
};

export interface ErrorResponse {
    traceID?: string;
    details?: string;
    fieldErrors?: FieldError[];
};

export interface UserRegistration {
    username: string;
    password: string;
};

export type AuthType = 'userPassword' | 'apiKey';

export interface Authentication {
    type: AuthType;
    username?: string;
    password?: string;
    key?: string;
};

export interface IndexerConfig {
    type: string;
    url: string;
    auth: Authentication;
};

export interface TransferClientConfig {
    type: string;
    url: string;
    auth: Authentication;
};

export interface Indexer {
    id: string;
    type: 'prowlarr';
    url: string;
};

export type WatchlistSource = 'local' | 'anilist' | 'mal';

export interface WatchlistConfig {
    source?: WatchlistSource;
    externalID: string;
    syncFrequencySeconds: number;
};