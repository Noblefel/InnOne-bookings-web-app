function validateDates(from, to) {
    from = luxon.DateTime.fromISO(from.value);
    to = luxon.DateTime.fromISO(to.value); 

    let days = to.diff(from, "days").values.days;  
 
    if(from.toISODate() < luxon.DateTime.now().toISODate()) { 
        return false
    }

    if(days < 0) { 
        return false
    }

    setStayingFor(days)

    return true
}

function calculateCost(days, costQuerySelector = '#payment') {  
 
    if (days == 0) days++; 

    setCost(defaultCost * days, costQuerySelector);

    return true
  }

function getDays(from, to) { 

    from = luxon.DateTime.fromISO(from.value);
    to = luxon.DateTime.fromISO(to.value);

    let days = to.diff(from, "days").values.days;

    return Math.max(0, days)
}

function setDefaultDates(from, to) {  

    from.value = luxon.DateTime.now().toISODate();
    to.value = luxon.DateTime.now().plus({ days: 1 }).toISODate();
    
    setStayingFor()
}

function setStayingFor(days = 1, querySelector = '#staying-for') { 
    document.querySelector(querySelector).innerText = days; 
}

function setCost(number, querySelector = "#payment") {
    let cost = new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(number);

    document.querySelector(querySelector).innerText = cost;
}
