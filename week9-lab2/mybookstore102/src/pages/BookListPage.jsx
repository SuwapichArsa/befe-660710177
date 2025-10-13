import React, { useState, useEffect } from 'react';
import BookCard from '../components/BookCard';
import SearchBar from '../components/SearchBar';
import LoadingSpinner from '../components/LoadingSpinner';
import { ChevronDownIcon } from '@heroicons/react/outline';

const BookListPage = () => {
  const [books, setBooks] = useState([]);
  const [filteredBooks, setFilteredBooks] = useState([]);
  const [sortBy, setSortBy] = useState('newest');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const booksPerPage = 12;
  const [error, setError] = useState(null);

  const categories = [
    'all', 'fiction', 'non-fiction', 'science', 'history', 'art',
    'psychology', 'business', 'technology', 'cooking'
  ];

  useEffect(() => {
    const fetchBooks = async () => {
      try {
        setLoading(true);

        // ✅ ใช้ .env หรือ fallback เป็น localhost
        const apiUrl = process.env.REACT_APP_API_URL || "http://localhost:8080";

        // ✅ ดึงข้อมูลจาก backend จริง (Go + PostgreSQL)
        const response = await fetch(`${apiUrl}/api/v1/books`);
        if (!response.ok) {
          throw new Error('Failed to fetch books');
        }

        const data = await response.json();

        // ✅ เก็บข้อมูลทั้งหมด
        setBooks(data);
        setFilteredBooks(data);
        setError(null);

      } catch (err) {
        setError(err.message);
        console.error('Error fetching books:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchBooks();
  }, []);

  // 🔍 Search
  const handleSearch = (searchTerm) => {
    const filtered = books.filter(book =>
      book.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      book.author.toLowerCase().includes(searchTerm.toLowerCase())
    );
    setFilteredBooks(filtered);
    setCurrentPage(1);
  };

  // 🔖 Category Filter
  const handleCategoryFilter = (category) => {
    setSelectedCategory(category);
    if (category === 'all') {
      setFilteredBooks(books);
    } else {
      const filtered = books.filter(book =>
        (book.category || '').toLowerCase() === category.toLowerCase()
      );
      setFilteredBooks(filtered);
    }
    setCurrentPage(1);
  };

  // 🔄 Sorting
  const handleSort = (sortValue) => {
    setSortBy(sortValue);
    const sorted = [...filteredBooks];
    switch (sortValue) {
      case 'price-low':
        sorted.sort((a, b) => a.price - b.price);
        break;
      case 'price-high':
        sorted.sort((a, b) => b.price - a.price);
        break;
      case 'popular':
        sorted.sort((a, b) => (b.reviews || 0) - (a.reviews || 0));
        break;
      default:
        sorted.sort((a, b) => b.id - a.id);
    }
    setFilteredBooks(sorted);
  };

  // 🔢 Pagination logic
  const indexOfLastBook = currentPage * booksPerPage;
  const indexOfFirstBook = indexOfLastBook - booksPerPage;
  const currentBooks = filteredBooks.slice(indexOfFirstBook, indexOfLastBook);
  const totalPages = Math.ceil(filteredBooks.length / booksPerPage);
  const paginate = (pageNumber) => setCurrentPage(pageNumber);

  if (loading) return <LoadingSpinner />;
  if (error) return <p className="text-center text-red-600 mt-10">เกิดข้อผิดพลาด: {error}</p>;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">หนังสือทั้งหมด</h1>
          <p className="text-gray-600">ค้นพบหนังสือที่คุณชื่นชอบจากคอลเล็กชันของเรา</p>
        </div>

        {/* Filters & Search */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-8">
          <div className="flex flex-col lg:flex-row gap-4">
            <div className="flex-1">
              <SearchBar onSearch={handleSearch} />
            </div>

            {/* Category */}
            <select
              className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none 
                focus:ring-2 focus:ring-viridian-500 cursor-pointer"
              value={selectedCategory}
              onChange={(e) => handleCategoryFilter(e.target.value)}
            >
              {categories.map((cat) => (
                <option key={cat} value={cat}>{cat}</option>
              ))}
            </select>

            {/* Sort */}
            <select
              className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none 
                focus:ring-2 focus:ring-viridian-500 cursor-pointer"
              value={sortBy}
              onChange={(e) => handleSort(e.target.value)}
            >
              <option value="newest">ใหม่ล่าสุด</option>
              <option value="price-low">ราคาต่ำ-สูง</option>
              <option value="price-high">ราคาสูง-ต่ำ</option>
              <option value="popular">ยอดนิยม</option>
            </select>
          </div>

          <div className="mt-4 text-sm text-gray-600">
            พบหนังสือ {filteredBooks.length} เล่ม
            {selectedCategory !== 'all' && ` ในหมวด ${selectedCategory}`}
          </div>
        </div>

        {/* Books */}
        {currentBooks.length > 0 ? (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {currentBooks.map(book => (
              <BookCard key={book.id} book={book} />
            ))}
          </div>
        ) : (
          <div className="text-center py-12">
            <p className="text-gray-500 text-lg">ไม่พบหนังสือที่ค้นหา</p>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="mt-12 flex justify-center">
            <nav className="flex items-center space-x-2">
              <button
                onClick={() => paginate(currentPage - 1)}
                disabled={currentPage === 1}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
              >
                ก่อนหน้า
              </button>

              {[...Array(totalPages)].map((_, index) => (
                <button
                  key={index + 1}
                  onClick={() => paginate(index + 1)}
                  className={`px-4 py-2 rounded-lg ${currentPage === index + 1
                    ? 'bg-viridian-600 text-white'
                    : 'border border-gray-300 hover:bg-gray-50'}`}
                >
                  {index + 1}
                </button>
              ))}

              <button
                onClick={() => paginate(currentPage + 1)}
                disabled={currentPage === totalPages}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
              >
                ถัดไป
              </button>
            </nav>
          </div>
        )}
      </div>
    </div>
  );
};

export default BookListPage;