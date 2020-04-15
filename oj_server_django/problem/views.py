from django.shortcuts import render
from .models import Problem


def problemlist(request):
    context = {'problem_list': Problem.objects.all()}
    return render(request, 'problem/list.html', context=context)


def search(request):
    if request.method == 'POST':  # 0 "" All All
        search_range = request.POST['search_range']
        search_string = request.POST['search_string']

    problemResObj = Problem.objects
    if search_string != None and search_string != "":
        if search_range == "problem_id":
            problemResObj = problemResObj.filter(id=search_string)
        if search_range == "title":
            problemResObj = problemResObj.filter(title=search_string)
        if search_range == "point":
            problemResObj = problemResObj.filter(point=search_string)
        context = {'problem_list': problemResObj,
                   'search_range': search_range, 'search_string': search_string}
    else:
        context = {'problem_list': problemResObj.all(),
                   'search_range': search_range, 'search_string': search_string}
    return render(request, 'problem/list.html', context=context)


def problem(request, problem_id):
    context = {
        'problem_list': Problem.objects.all(),
        'problem_id': problem_id
    }
    return render(request, 'problem/problem.html', context=context)


def submit(request, problem_id):
    context = {'problem_id': problem_id}
    return render(request, 'problem/submit.html', context=context)
