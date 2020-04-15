State = {
    name="",
    age=0,
    child = {
        name=""
    }
}

function State:new(name, age, subObjName)
    State.name = name
    State.age = age
    State.child = {
        name=subObjName
    }
end

function getName()
    return State.name, State.age
end

function getAge()
    return State.age
end

function getChildObject()
    return State.child
end