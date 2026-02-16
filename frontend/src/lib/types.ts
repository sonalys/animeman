export interface UserRegistration {
    username: string;
    password: string;
}

export interface AuthResponse {
    userID: string;
}

export interface FieldError {
    field: string;
    code: 'alreadyExists' | 'minLength' | 'maxLength' | 'required' | 'invalidFormat' | 'unknown';
    message: string;
}

export interface ErrorResponse {
    traceID?: string;
    details?: string;
    fieldErrors?: FieldError[];
}

export interface UserRegistration {
    username: string; // pattern: ^[a-zA-Z0-9_]+$
    password: string; // min: 8, max: 72
}

export type AuthType = 'userPassword' | 'apiKey';

export interface Authentication {
    type: AuthType;
    username?: string;
    password?: string;
    key?: string;
}

export interface IndexerConfig {
    type: string;
    url: string;
    auth: Authentication;
}

export interface Indexer {
    id: string;
    type: 'prowlarr' | 'jackett' | 'torznab';
    url: string;
    status: 'online' | 'offline' | 'unauthorized';
    // We don't usually return the sensitive auth data back to the UI
}