State = {
    name="",
    age=0,
}

function State:new(name, age)

    local obj= {}
    obj.name = name
    obj.age = age

    function obj:getName()
        print(self.name)
        return self.name
    end

    function obj:getAge()
        return self.age
    end

    setmetatable(obj, self)
    self.__index = self

    return obj
end

--vasya = State:new("John", 12)
--print(vasya:getName())
--print(vasya:getAge())