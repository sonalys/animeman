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

export interface AuthenticationUserPassword {
    type: 'userPassword'
    username: string;
    password: string;
}

export interface AuthenticationAPIKey {
    type: 'apiKey';
    key: string;
}

export interface AuthenticationNone {
    type: 'none';
}

export type Authentication = AuthenticationUserPassword | AuthenticationAPIKey | AuthenticationNone;

export interface IndexerConfig {
    type: string;
    hostname: string;
    auth: Authentication;
};

export interface TransferClientConfig {
    type: 'qbittorrent'
    hostname: string;
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