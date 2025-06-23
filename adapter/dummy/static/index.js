let self = fetchObj('/api/self')

function selfInfo() {
    return Object.assign({}, self)
}

let users = fetchObj('/api/users')

function usersInfo() {
    return Object.assign({}, users)
}

function getSession(id) {
    return getSessions()[id];
}

function getSessions() {
    return fetchObj('/api/sessions');
}

function getUserInfo(id) {
    let user = usersInfo()[id];
    if (user == null) {
        return null
    }
    user['id'] = id;
    return user;
}

function getGroupInfo(id) {
    return fetchObj('/api/groups')[id]
}


function getMessages(id) {
    let msg = fetchObj('/api/messages?sid=' + id)
    for (let item of msg) {
        let u = getUserInfo(item.sender)
        item.name = u.name
        item.avatar = u.avatar
        item.created = timeFormat(item.created)
        item.isSelf = u.id === selfInfo().id
    }
    let session = getSession(id);
    if (session.type === "group") {
        let groupInfo = getGroupInfo(session.target);
        for (const item of msg) {
            for (let uid of Object.keys(groupInfo.users)) {
                if (uid === item.sender) {
                    if (groupInfo.users[uid].name) {
                        item.name = groupInfo.users[uid].name
                    }
                    break
                }
            }
        }
    }
    return msg;
}


function getMessage(id) {
    return fetchObj('/api/message?id=' + id);
}

function searchUser(sessionID, matchUser) {
    let match = matchUser.toLowerCase()
    let result = []
    const session = getSession(sessionID)
    if (session.type === 'group') {
        let groupInfo = getGroupInfo(session.target);
        for (let [uid, meta] of Object.entries(groupInfo.users)) {
            const uu = getUserInfo(uid)
            if (
                (meta.name && meta.name.toLowerCase().includes(match)) ||
                uu.name.toLowerCase().includes(match) ||
                uu.id.includes(matchUser)
            ) {
                if (meta.name) {
                    uu.groupName = meta.name
                }
                result.push(uu)
            }
        }
    }
    return result
}

