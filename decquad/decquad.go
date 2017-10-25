package decquad

import (
	"errors"
)

// The decQuad decimal 128-bit type, accessible by all sizes
type DecQuad [4]uint32 // little-endian

// Context for operations, with associated constants
type Rounding byte

const (
	_                   Rounding = iota
	DEC_ROUND_CEILING            // round towards +infinity
	DEC_ROUND_UP                 // round away from 0
	DEC_ROUND_HALF_UP            // 0.5 rounds up
	DEC_ROUND_HALF_EVEN          // 0.5 rounds to nearest even
	DEC_ROUND_HALF_DOWN          // 0.5 rounds down
	DEC_ROUND_DOWN               // round towards 0 (truncate)
	DEC_ROUND_FLOOR              // round towards -infinity
	DEC_ROUND_05UP               // round for reround
	DEC_ROUND_MAX                // enum must be less than this
)

var DEC_ROUND_DEFAULT Rounding = DEC_ROUND_HALF_UP

type DecContext struct {
	digits   int32    // working precision
	emax     int32    // maximum positive exponent
	emin     int32    // minimum negative exponent
	round    Rounding // rounding mode
	traps    uint32   // trap-enabler flags
	status   uint32   // status flags
	clamp    uint8    // flag: apply IEEE exponent clamp
	extended uint8    // flag: special-values allowed
}

const (
	// sign and special values [top 32-bits; last two bits are don't-care
	//for Infinity on input, last bit don't-care for NaNs]
	DECFLOAT_Sign  = 0x80000000 // 1 00000 00 Sign
	DECFLOAT_NaN   = 0x7c000000 // 0 11111 00 NaN generic
	DECFLOAT_qNaN  = 0x7c000000 // 0 11111 00 qNaN
	DECFLOAT_sNaN  = 0x7e000000 // 0 11111 10 sNaN
	DECFLOAT_Inf   = 0x78000000 // 0 11110 00 Infinity
	DECFLOAT_MinSp = 0x78000000 // minimum special value [specials are all >=MinSp]

	DEC_Conversion_syntax    = 0x00000001
	DEC_Division_by_zero     = 0x00000002
	DEC_Division_impossible  = 0x00000004
	DEC_Division_undefined   = 0x00000008
	DEC_Insufficient_storage = 0x00000010 // [when malloc fails]
	DEC_Inexact              = 0x00000020
	DEC_Invalid_context      = 0x00000040
	DEC_Invalid_operation    = 0x00000080
	DEC_Lost_digits          = 0x00000100
	DEC_Overflow             = 0x00000200
	DEC_Clamped              = 0x00000400
	DEC_Rounded              = 0x00000800
	DEC_Subnormal            = 0x00001000
	DEC_Underflow            = 0x00002000
)

var DECCOMBFROM = [48]uint32{
	0x00000000, 0x04000000, 0x08000000, 0x0C000000, 0x10000000, 0x14000000,
	0x18000000, 0x1C000000, 0x60000000, 0x64000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000, 0x20000000, 0x24000000,
	0x28000000, 0x2C000000, 0x30000000, 0x34000000, 0x38000000, 0x3C000000,
	0x68000000, 0x6C000000, 0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x40000000, 0x44000000, 0x48000000, 0x4C000000,
	0x50000000, 0x54000000, 0x58000000, 0x5C000000, 0x70000000, 0x74000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000, 0x00000000, 0x00000000}

var BCD2DPD = [2458]uint32{0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 0, 0, 0, 0, 0, 0, 16, 17, 18, 19, 20,
	21, 22, 23, 24, 25, 0, 0, 0, 0, 0, 0, 32, 33,
	34, 35, 36, 37, 38, 39, 40, 41, 0, 0, 0, 0, 0,
	0, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 0, 0,
	0, 0, 0, 0, 64, 65, 66, 67, 68, 69, 70, 71, 72,
	73, 0, 0, 0, 0, 0, 0, 80, 81, 82, 83, 84, 85,
	86, 87, 88, 89, 0, 0, 0, 0, 0, 0, 96, 97, 98,
	99, 100, 101, 102, 103, 104, 105, 0, 0, 0, 0, 0, 0,
	112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 0, 0, 0,
	0, 0, 0, 10, 11, 42, 43, 74, 75, 106, 107, 78, 79,
	0, 0, 0, 0, 0, 0, 26, 27, 58, 59, 90, 91, 122,
	123, 94, 95, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 0, 0,
	0, 0, 0, 0, 144, 145, 146, 147, 148, 149, 150, 151, 152,
	153, 0, 0, 0, 0, 0, 0, 160, 161, 162, 163, 164, 165,
	166, 167, 168, 169, 0, 0, 0, 0, 0, 0, 176, 177, 178,
	179, 180, 181, 182, 183, 184, 185, 0, 0, 0, 0, 0, 0,
	192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 0, 0, 0,
	0, 0, 0, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217,
	0, 0, 0, 0, 0, 0, 224, 225, 226, 227, 228, 229, 230,
	231, 232, 233, 0, 0, 0, 0, 0, 0, 240, 241, 242, 243,
	244, 245, 246, 247, 248, 249, 0, 0, 0, 0, 0, 0, 138,
	139, 170, 171, 202, 203, 234, 235, 206, 207, 0, 0, 0, 0,
	0, 0, 154, 155, 186, 187, 218, 219, 250, 251, 222, 223, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 256, 257, 258,
	259, 260, 261, 262, 263, 264, 265, 0, 0, 0, 0, 0, 0,
	272, 273, 274, 275, 276, 277, 278, 279, 280, 281, 0, 0, 0,
	0, 0, 0, 288, 289, 290, 291, 292, 293, 294, 295, 296, 297,
	0, 0, 0, 0, 0, 0, 304, 305, 306, 307, 308, 309, 310,
	311, 312, 313, 0, 0, 0, 0, 0, 0, 320, 321, 322, 323,
	324, 325, 326, 327, 328, 329, 0, 0, 0, 0, 0, 0, 336,
	337, 338, 339, 340, 341, 342, 343, 344, 345, 0, 0, 0, 0,
	0, 0, 352, 353, 354, 355, 356, 357, 358, 359, 360, 361, 0,
	0, 0, 0, 0, 0, 368, 369, 370, 371, 372, 373, 374, 375,
	376, 377, 0, 0, 0, 0, 0, 0, 266, 267, 298, 299, 330,
	331, 362, 363, 334, 335, 0, 0, 0, 0, 0, 0, 282, 283,
	314, 315, 346, 347, 378, 379, 350, 351, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 384, 385, 386, 387, 388, 389, 390,
	391, 392, 393, 0, 0, 0, 0, 0, 0, 400, 401, 402, 403,
	404, 405, 406, 407, 408, 409, 0, 0, 0, 0, 0, 0, 416,
	417, 418, 419, 420, 421, 422, 423, 424, 425, 0, 0, 0, 0,
	0, 0, 432, 433, 434, 435, 436, 437, 438, 439, 440, 441, 0,
	0, 0, 0, 0, 0, 448, 449, 450, 451, 452, 453, 454, 455,
	456, 457, 0, 0, 0, 0, 0, 0, 464, 465, 466, 467, 468,
	469, 470, 471, 472, 473, 0, 0, 0, 0, 0, 0, 480, 481,
	482, 483, 484, 485, 486, 487, 488, 489, 0, 0, 0, 0, 0,
	0, 496, 497, 498, 499, 500, 501, 502, 503, 504, 505, 0, 0,
	0, 0, 0, 0, 394, 395, 426, 427, 458, 459, 490, 491, 462,
	463, 0, 0, 0, 0, 0, 0, 410, 411, 442, 443, 474, 475,
	506, 507, 478, 479, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 512, 513, 514, 515, 516, 517, 518, 519, 520, 521, 0,
	0, 0, 0, 0, 0, 528, 529, 530, 531, 532, 533, 534, 535,
	536, 537, 0, 0, 0, 0, 0, 0, 544, 545, 546, 547, 548,
	549, 550, 551, 552, 553, 0, 0, 0, 0, 0, 0, 560, 561,
	562, 563, 564, 565, 566, 567, 568, 569, 0, 0, 0, 0, 0,
	0, 576, 577, 578, 579, 580, 581, 582, 583, 584, 585, 0, 0,
	0, 0, 0, 0, 592, 593, 594, 595, 596, 597, 598, 599, 600,
	601, 0, 0, 0, 0, 0, 0, 608, 609, 610, 611, 612, 613,
	614, 615, 616, 617, 0, 0, 0, 0, 0, 0, 624, 625, 626,
	627, 628, 629, 630, 631, 632, 633, 0, 0, 0, 0, 0, 0,
	522, 523, 554, 555, 586, 587, 618, 619, 590, 591, 0, 0, 0,
	0, 0, 0, 538, 539, 570, 571, 602, 603, 634, 635, 606, 607,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 640, 641,
	642, 643, 644, 645, 646, 647, 648, 649, 0, 0, 0, 0, 0,
	0, 656, 657, 658, 659, 660, 661, 662, 663, 664, 665, 0, 0,
	0, 0, 0, 0, 672, 673, 674, 675, 676, 677, 678, 679, 680,
	681, 0, 0, 0, 0, 0, 0, 688, 689, 690, 691, 692, 693,
	694, 695, 696, 697, 0, 0, 0, 0, 0, 0, 704, 705, 706,
	707, 708, 709, 710, 711, 712, 713, 0, 0, 0, 0, 0, 0,
	720, 721, 722, 723, 724, 725, 726, 727, 728, 729, 0, 0, 0,
	0, 0, 0, 736, 737, 738, 739, 740, 741, 742, 743, 744, 745,
	0, 0, 0, 0, 0, 0, 752, 753, 754, 755, 756, 757, 758,
	759, 760, 761, 0, 0, 0, 0, 0, 0, 650, 651, 682, 683,
	714, 715, 746, 747, 718, 719, 0, 0, 0, 0, 0, 0, 666,
	667, 698, 699, 730, 731, 762, 763, 734, 735, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 768, 769, 770, 771, 772, 773,
	774, 775, 776, 777, 0, 0, 0, 0, 0, 0, 784, 785, 786,
	787, 788, 789, 790, 791, 792, 793, 0, 0, 0, 0, 0, 0,
	800, 801, 802, 803, 804, 805, 806, 807, 808, 809, 0, 0, 0,
	0, 0, 0, 816, 817, 818, 819, 820, 821, 822, 823, 824, 825,
	0, 0, 0, 0, 0, 0, 832, 833, 834, 835, 836, 837, 838,
	839, 840, 841, 0, 0, 0, 0, 0, 0, 848, 849, 850, 851,
	852, 853, 854, 855, 856, 857, 0, 0, 0, 0, 0, 0, 864,
	865, 866, 867, 868, 869, 870, 871, 872, 873, 0, 0, 0, 0,
	0, 0, 880, 881, 882, 883, 884, 885, 886, 887, 888, 889, 0,
	0, 0, 0, 0, 0, 778, 779, 810, 811, 842, 843, 874, 875,
	846, 847, 0, 0, 0, 0, 0, 0, 794, 795, 826, 827, 858,
	859, 890, 891, 862, 863, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 896, 897, 898, 899, 900, 901, 902, 903, 904, 905,
	0, 0, 0, 0, 0, 0, 912, 913, 914, 915, 916, 917, 918,
	919, 920, 921, 0, 0, 0, 0, 0, 0, 928, 929, 930, 931,
	932, 933, 934, 935, 936, 937, 0, 0, 0, 0, 0, 0, 944,
	945, 946, 947, 948, 949, 950, 951, 952, 953, 0, 0, 0, 0,
	0, 0, 960, 961, 962, 963, 964, 965, 966, 967, 968, 969, 0,
	0, 0, 0, 0, 0, 976, 977, 978, 979, 980, 981, 982, 983,
	984, 985, 0, 0, 0, 0, 0, 0, 992, 993, 994, 995, 996,
	997, 998, 999, 1000, 1001, 0, 0, 0, 0, 0, 0, 1008, 1009,
	1010, 1011, 1012, 1013, 1014, 1015, 1016, 1017, 0, 0, 0, 0, 0,
	0, 906, 907, 938, 939, 970, 971, 1002, 1003, 974, 975, 0, 0,
	0, 0, 0, 0, 922, 923, 954, 955, 986, 987, 1018, 1019, 990,
	991, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12,
	13, 268, 269, 524, 525, 780, 781, 46, 47, 0, 0, 0, 0,
	0, 0, 28, 29, 284, 285, 540, 541, 796, 797, 62, 63, 0,
	0, 0, 0, 0, 0, 44, 45, 300, 301, 556, 557, 812, 813,
	302, 303, 0, 0, 0, 0, 0, 0, 60, 61, 316, 317, 572,
	573, 828, 829, 318, 319, 0, 0, 0, 0, 0, 0, 76, 77,
	332, 333, 588, 589, 844, 845, 558, 559, 0, 0, 0, 0, 0,
	0, 92, 93, 348, 349, 604, 605, 860, 861, 574, 575, 0, 0,
	0, 0, 0, 0, 108, 109, 364, 365, 620, 621, 876, 877, 814,
	815, 0, 0, 0, 0, 0, 0, 124, 125, 380, 381, 636, 637,
	892, 893, 830, 831, 0, 0, 0, 0, 0, 0, 14, 15, 270,
	271, 526, 527, 782, 783, 110, 111, 0, 0, 0, 0, 0, 0,
	30, 31, 286, 287, 542, 543, 798, 799, 126, 127, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 140, 141, 396, 397, 652,
	653, 908, 909, 174, 175, 0, 0, 0, 0, 0, 0, 156, 157,
	412, 413, 668, 669, 924, 925, 190, 191, 0, 0, 0, 0, 0,
	0, 172, 173, 428, 429, 684, 685, 940, 941, 430, 431, 0, 0,
	0, 0, 0, 0, 188, 189, 444, 445, 700, 701, 956, 957, 446,
	447, 0, 0, 0, 0, 0, 0, 204, 205, 460, 461, 716, 717,
	972, 973, 686, 687, 0, 0, 0, 0, 0, 0, 220, 221, 476,
	477, 732, 733, 988, 989, 702, 703, 0, 0, 0, 0, 0, 0,
	236, 237, 492, 493, 748, 749, 1004, 1005, 942, 943, 0, 0, 0,
	0, 0, 0, 252, 253, 508, 509, 764, 765, 1020, 1021, 958, 959,
	0, 0, 0, 0, 0, 0, 142, 143, 398, 399, 654, 655, 910,
	911, 238, 239, 0, 0, 0, 0, 0, 0, 158, 159, 414, 415,
	670, 671, 926, 927, 254, 255}

type bcdnum struct {
	msd      [64]byte // -> most significant digit
	lsd      int      /// -> least ditto
	sign     uint32   // 0=positive, DECFLOAT_Sign=negative
	exponent int32    // Unadjusted signed exponent (q), or DECFLOAT_NaN etc. for a special
}

func DecFloatFromString(s string, set *DecContext) (DecQuad, error) {

	var retval DecQuad

	if len(s) == 0 {
		return retval, errors.New("Пустая строка недопустима")
	}

	var num bcdnum

	if s == "Inf" {
		num.exponent = DECFLOAT_Inf
	}
	if s == "NaN" {
		num.exponent = DECFLOAT_qNaN
	}

	var p, pPoint, pE, lenexp int
	var exp int32
	var negE bool

	for i, ch := range s {
		if p >= 64 {
			return retval, errors.New("Слишком длинное число")
		}
		edig := ch - '0'
		switch {
		case edig <= 9:
			if pE == 0 {
				if p == 0 && edig == 0 {
					continue
				}
				num.msd[p] = byte(edig)
				p++
			} else {
				if exp == 0 && edig == 0 {
					continue
				}
				// собираем экспоненту
				exp = exp*10 + edig
				lenexp++
			}
		case ch == '.':
			if i == 0 {
				return retval, errors.New("Число должно начинаться с цифры")
			}
			if pPoint != 0 {
				return retval, errors.New("Несколько точек недопустимо")
			}
			if pE != 0 {
				return retval, errors.New("Точка в экспоненте недопустима")
			}
			pPoint = p
			// если до точки все были нулями - двигаем указатель (а ноль уже и так будет сохранен)
			if p == 0 {
				p++
			}
		case i == 0 && ch == '-':
			num.sign = DECFLOAT_Sign
		case (ch == 'e' || ch == 'E') && pE == 0:
			pE = p
		case pE == p && ch == '-':
			negE = true
		default:
			return retval, errors.New("Неверный формат числа")
		}

	}
	num.lsd = p
	if lenexp > 4 {
		exp = 6144
	}
	if pPoint > 0 && pE > pPoint {
		exp -= int32(pE - pPoint) //1.5E2 = 15E1
	}
	if negE {
		exp = -exp
	}
	num.exponent = exp

	retval = decFinalize(num, set)

	return retval, nil
}

func decFinalize(num bcdnum, set *DecContext) (df DecQuad) {
	var encode uint32
	if num.exponent < DECFLOAT_MinSp {
		uexp := uint32(num.exponent + 6176) // biased exponent
		code := (uexp >> 12) << 4           // top two bits of exp
		// [msd==0]
		// look up the combination field and make high word
		encode = DECCOMBFROM[code]                     // indexed by (0-2)*16+msd
		encode |= (uexp << (32 - 6 - 12)) & 0x03ffffff // exponent continuation

	} else {
		encode = uint32(num.exponent) // special [already in word]
	}
	encode |= num.sign // add sign

	var n int
	var dpd uint32

	n = 10
	ub := num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) << 4

	n = 9
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) >> 6

	df[3] = encode

	encode = uint32(dpd) << 26

	n = 8
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) << 16

	n = 7
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) << 6

	n = 6
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) >> 4
	df[2] = encode

	encode = uint32(dpd) << 28

	n = 5
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) << 18

	n = 4
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) << 8

	n = 3
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) >> 2
	df[1] = encode

	encode = uint32(dpd) << 30

	n = 2
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}
	encode |= uint32(dpd) << 20

	n = 1
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= uint32(dpd) << 10

	n = 0
	ub = num.lsd - (3 * n) - 2
	if ub < (-2) {
		dpd = 0
	} else if ub >= 0 {
		dpd = BCD2DPD[(int(num.msd[ub])*256)+(int(num.msd[ub+1])*16)+int(num.msd[ub+2])]
	} else {
		dpd = uint32(num.msd[ub+2])
		if ub+1 == 0 {
			dpd += uint32(num.msd[ub+1]) * 16
		}
		dpd = BCD2DPD[dpd]
	}

	encode |= dpd

	df[0] = encode

	return
}
