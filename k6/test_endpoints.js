import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Trend } from 'k6/metrics';

export let options = {
  vus: __ENV.K6_VUS ? parseInt(__ENV.K6_VUS) : 100,
  duration: __ENV.K6_DURATION || '30s',
  thresholds: {
    'http_req_duration{endpoint:books_list}': ['p(95)<200'],
    'http_req_duration{endpoint:roles_create}': ['p(95)<300'],
    'http_req_duration{endpoint:contacts_list}': ['p(95)<200'],
    'http_req_failed': ['rate<0.01'],
  },
};

const BASE = __ENV.K6_BASE_URL || 'http://localhost:3000'

function mkParams(token, endpoint) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;
  return { headers, tags: { endpoint } };
}

// custom trends
const books_duration = new Trend('books_duration');
const profile_duration = new Trend('profile_duration');
const roles_create_duration = new Trend('roles_create_duration');
const roles_patch_duration = new Trend('roles_patch_duration');
const contacts_duration = new Trend('contacts_duration');

export function setup() {
  // prefer an explicit token, otherwise perform login with K6_USER/K6_PASS
  const tokenEnv = __ENV.K6_TOKEN;
  if (tokenEnv) return { token: tokenEnv };

  const user = __ENV.K6_USER || 'amiana@gmail.com';
  const pass = __ENV.K6_PASS || 'amiana';
  const res = http.post(`${BASE}/api/login`, JSON.stringify({ email: user, password: pass }), mkParams('', 'auth_login'));
  if (res.status !== 200) {
    throw new Error(`K6 setup: login failed status=${res.status} body=${res.body}`);
  }
  const body = (() => { try { return res.json(); } catch (e) { return {}; } })();
  const token = body.token || body.access_token || '';
  if (!token) {
    throw new Error('K6 setup: no token obtained from login or K6_TOKEN');
  }
  return { token };
}

export default function (data) {
  const token = data.token || '';
  if (!token) {
    throw new Error('No token provided to VU; aborting');
  }

  function doCheck(res, checks, name) {
    const ok = check(res, checks);
    if (!ok) {
      // log status and body for debugging
      try {
        console.error(`${name} failed: status=${res.status} body=${res.body}`);
      } catch (e) {
        console.error(`${name} failed and response body could not be read`);
      }
    }
    return ok;
  }

  group('Public endpoints', function () {
    const params = mkParams(token, 'books_list');
    const r1 = http.get(`${BASE}/api/books`, params);
    books_duration.add(r1.timings.duration);
    doCheck(r1, { 'books 200': (r) => r.status === 200 }, 'books_list');
  });

  group('Protected profile', function () {
    const params = mkParams(token, 'profile_get');
    const r = http.get(`${BASE}/api/profile`, params);
    profile_duration.add(r.timings.duration);
    doCheck(r, { 'profile 200': (r) => r.status === 200 }, 'profile_get');
  });

  group('Roles CRUD (create -> patch)', function () {
    const createParams = mkParams(token, 'roles_create');
    const create = http.post(`${BASE}/api/roles`, JSON.stringify({ name: `role_${Math.floor(Math.random() * 10000)}` }), createParams);
    roles_create_duration.add(create.timings.duration);
    doCheck(create, { 'role create 2xx': (r) => r.status >= 200 && r.status < 300 }, 'roles_create');
    let roleID = '';
    try { roleID = create.json().id } catch (e) {}
    if (roleID) {
      const patchParams = mkParams(token, 'roles_patch');
      const patch = http.patch(`${BASE}/api/roles/${roleID}`, JSON.stringify({ name: `role_updated_${Math.floor(Math.random() * 10000)}` }), patchParams);
      roles_patch_duration.add(patch.timings.duration);
      doCheck(patch, { 'role patch 2xx': (r) => r.status >= 200 && r.status < 300 }, 'roles_patch');
    }
  });

  group('Sites CRUD (create -> patch)', function () {
    const createParams = mkParams(token, 'sites_create');
    const create = http.post(`${BASE}/api/sites`, JSON.stringify({ name: `site_${Math.floor(Math.random() * 10000)}` }), createParams);
    doCheck(create, { 'site create 2xx': (r) => r.status >= 200 && r.status < 300 }, 'sites_create');
    let siteUUID = '';
    try { siteUUID = create.json().uuid } catch (e) {}
    if (siteUUID) {
      const patchParams = mkParams(token, 'sites_patch');
      const patch = http.patch(`${BASE}/api/sites/${siteUUID}`, JSON.stringify({ name: `site_updated_${Math.floor(Math.random() * 10000)}` }), patchParams);
      doCheck(patch, { 'site patch 2xx': (r) => r.status >= 200 && r.status < 300 }, 'sites_patch');
    }
  });

  group('Contacts CRUD (create -> patch -> list)', function () {
    const payload = { Name: `Contact ${Math.floor(Math.random() * 10000)}`, Phone: '081234567890' };
    const createParams = mkParams(token, 'contacts_create');
    const create = http.post(`${BASE}/api/contacts`, JSON.stringify(payload), createParams);
    doCheck(create, { 'contact create 2xx': (r) => r.status >= 200 && r.status < 300 }, 'contacts_create');
    let contactID = '';
    try { contactID = create.json().Id || create.json().id } catch (e) {}
    if (contactID) {
      const patchParams = mkParams(token, 'contacts_patch');
      const patch = http.patch(`${BASE}/api/contacts/${contactID}`, JSON.stringify({ Name: `Contact Updated ${Math.floor(Math.random() * 10000)}` }), patchParams);
      doCheck(patch, { 'contact patch 2xx': (r) => r.status >= 200 && r.status < 300 }, 'contacts_patch');
    }
    const listParams = mkParams(token, 'contacts_list');
    const list = http.get(`${BASE}/api/contacts`, listParams);
    contacts_duration.add(list.timings.duration);
    doCheck(list, { 'contacts list 200': (r) => r.status === 200 }, 'contacts_list');
  });

  group('Bible endpoints (books, chapters, verses)', function () {
    // list books
    const booksParams = mkParams(token, 'books_list');
    const books = http.get(`${BASE}/api/books`, booksParams);
    books_duration.add(books.timings.duration);
    doCheck(books, { 'books list 200': (r) => r.status === 200 }, 'books_list');

    // try to pick first book and first chapter if available
    let bookId = '';
    try { const b = books.json(); if (Array.isArray(b) && b.length) bookId = b[0].id || b[0].ID } catch (e) {}
    if (bookId) {
      const chaptersParams = mkParams(token, 'chapters_list');
      const chapters = http.get(`${BASE}/api/books/${bookId}/chapters`, chaptersParams);
      doCheck(chapters, { 'chapters list 200': (r) => r.status === 200 }, 'chapters_list');
      try {
        const ch = chapters.json();
        if (Array.isArray(ch) && ch.length) {
          const chapterNum = ch[0].nomor_pasal || ch[0].NomorPasal || ch[0].nomor || ch[0].Nomor || ch[0].id;
          // If API expects /books/:book_id/chapters/:id where :id is nomor_pasal
          const pasalParams = mkParams(token, 'pasal_detail');
          const pasal = http.get(`${BASE}/api/books/${bookId}/chapters/${chapterNum}`, pasalParams);
          doCheck(pasal, { 'pasal detail 200': (r) => r.status === 200 }, 'pasal_detail');
          // verses via chapter id (if pasal returned contains id)
          const pasalBody = (() => { try { return pasal.json(); } catch (e) { return {}; } })();
          const pasalID = pasalBody.Pasal && pasalBody.Pasal.ID ? pasalBody.Pasal.ID : pasalBody.Pasal && pasalBody.Pasal.id ? pasalBody.Pasal.id : null;
          if (pasalID) {
            const versesParams = mkParams(token, 'verses_list');
            const verses = http.get(`${BASE}/api/chapters/${pasalID}/verses`, versesParams);
            doCheck(verses, { 'verses list 200': (r) => r.status === 200 }, 'verses_list');
            // get a verse detail
            try {
              const v = verses.json(); if (Array.isArray(v) && v.length) {
                const verseID = v[0].id || v[0].ID;
                if (verseID) {
                  const verseParams = mkParams(token, 'verse_detail');
                  const verse = http.get(`${BASE}/api/verses/${verseID}`, verseParams);
                  doCheck(verse, { 'verse detail 200': (r) => r.status === 200 }, 'verse_detail');
                }
              }
            } catch (e) {}
          }
        }
      } catch (e) {}
    }
  });

  sleep(Math.random() * 3);
}
