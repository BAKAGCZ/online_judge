from django.urls import path
from . import views
urlpatterns = [
    path('<int:pindex>/', views.status, name='status'),
    path('showsource/<int:solution_id>/', views.showsource, name='showsource'),
    path('submithandler/', views.submithandler, name='submithandler'),
    path('search/<int:pindex>/', views.search, name='search'),
]
