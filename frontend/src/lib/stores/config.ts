// src/lib/stores/config.ts
import { writable } from 'svelte/store';

export const configStore = writable({
    indexer: {
        type: 'Prowlarr',
        url: 'http://localhost:9696',
        apiKey: ''
    },
    transfer: {
        type: 'qBittorrent',
        url: 'http://localhost:8080',
        username: '',
        password: ''
    },
    collection: {
        name: 'My Anime Collection',
        path: '/volume1/media/anime',
        qualityProfile: {
            name: 'FullHD only',
            minRes: '1080p',
            maxRes: '1080p'
        }
    },
    watchlist: {
        source: 'AniList',
        username: ''
    }
});