local contract = require("contract")

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
    res = contract.call()
    print(res)
    return State.name, State.age
end

function getAge()
    return State.age
end

function getChildObject()
    return State.child
end