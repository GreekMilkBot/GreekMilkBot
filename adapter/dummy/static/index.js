let users = {}
let groups = {}
let sessions = {}
let self = {}
users = fetchObj('/api/users')
groups = fetchObj('/api/groups')
sessions = fetchObj('/api/sessions')
self = fetchObj('/api/self')


function selfInfo() {
    return self
}

function getSession(id) {
    return sessions[id];
}

function getSessions() {
    return sessions;
}

function getUserInfo(id) {
    let user = users[id];
    if (user == null) {
        return null
    }
    user['id'] = id;
    return user;
}

function getGroupInfo(id) {
    return groups[id]
}


function getMessages(id) {
    let msg = fetchObj('/api/message?sid=' + id)
    for (let item of msg) {
        let u = getUserInfo(item.sender)
        item.name = u.name
        item.avatar = u.avatar
        item.created = timeFormat(item.created)
        item.isSelf = u.id === selfInfo().id
    }
    return msg;
}


function getMessage(sid, mid) {
    const message = fetchObj('/api/message?sid=' + sid);
    for (let data of message) {
        if (data.id === mid) {
            return data
        }
    }
    return null
}

function searchUser(sessionID, matchUser) {
    let result = []
    if (sessions[sessionID].type === 'group') {
        groups[sessions[sessionID].target].users.forEach(user => {
            const uu = getUserInfo(user.uid)
            if (user.name.toLowerCase().includes(matchUser) || uu.name.toLowerCase().includes(matchUser) || uu.id.includes(matchUser)) {
                if (user.name !== "") {
                    uu.name = user.name
                }
                result.push(uu)
            }
        })
    }
    return result
}

