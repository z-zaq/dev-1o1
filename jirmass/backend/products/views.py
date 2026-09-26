from rest_framework import generics
from .models import Product
from .serializers import ProductSerializer

class ProductListCreateAPIView(generics.ListCreateAPIView):
    queryset = Product.objects.filter(is_active=True)
    serializer_class = ProductSerializer
