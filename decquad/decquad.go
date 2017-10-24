package decquad

import (
	"io"
	"strings"

)

// The decQuad decimal 128-bit type, accessible by all sizes
type decQuad [2]int64

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

type bcdnum struct {
	msd      [64]byte // -> most significant digit
	lsd      []byte   /// -> least ditto
	sign     uint32   // 0=positive, DECFLOAT_Sign=negative
	exponent int32    // Unadjusted signed exponent (q), or DECFLOAT_NaN etc. for a special
}

func DecFloatFromString(s string) decQuad {
	r:=strings.NewReader(s)

	var digits int
	dotchar := -1
	cfirst := 0
	var ub int
	var uiwork uint32
	var num bcdnum
	var err = DEC_Conversion_syntax
	var buf [64]byte
	var ic int

	for { // once-only 'loop'
		num.sign = 0  // assume non-negative
		num.msd = buf // MSD is here always

		// detect and validate the coefficient, including any leading,
		// trailing, or embedded '.'
		// [could test four-at-a-time here (saving 10% for decQuads),
		// but that risks storage violation because the position of the
		// terminator is unknown]
		r.Reset()
		ic=0
		lc:=0
		for c,cs,ce:=r.ReadRune();ce!=io.EOF;c,cs,ce=r.ReadRune(){ // -> input character
			lc=ic
			ic+=cs
			if (c - '0') <= 9 {
				continue
			} // '0' through '9' is good
			if c == '.' {
				if dotchar != -1 {
					break
				} // not first '.'
				dotchar = lc // record offset into decimal part
				continue
			}
			if i == 0 { // first in string...
				if c == '-' { // valid - sign
					cfirst++
					num.sign = DECFLOAT_Sign
					continue
				}
				if c == '+' { // valid + sign
					cfirst++
					continue
				}
			}
			// *c is not a digit, terminator, or a valid +, -, or '.'
			break
		} // c loop
	

		digits=ic-cfirst            // digits (+1 if a dot)
	
		if digits>0 {                     // had digits and/or dot
		  clast:=lc;            // note last coefficient char position
		  exp:=0;                        // exponent accumulator
		  if ic<len(s) {                   // something follows the coefficient
			var edig uint                      // unsigned work
			// had some digits and more to come; expect E[+|-]nnn now
			var firstexp int           // exponent first non-zero
			
			if c,cs,ce:=r.ReadRune();(ce==io.EOF) || (c!='E' && c!='e') {break}else{ic+=cs}
			c,cs,ce:=r.ReadRune()// to (optional) sign
			ic+=cs
			if ce==io.EOF{break}
			sneg:=c=='-'
			if c=='-' || c=='+' {c,cs,ce=r.ReadRune();ic+=cs}    // step over sign (c=clast+2)
			if ce==io.EOF{break}            // no digits!  (e.g., '1.2E')
			for ; ce!=io.EOF && c=='0'; c,cs,ce=r.ReadRune(){ic+=cs}           // skip leading zeros [even last]
			firstexp=ic;                     // remember start [maybe '\0']
			// gather exponent digits
			ndigs:=0
			edig=c-'0'
			if edig<=9 {                  // [check not bad or terminator]
			  exp+=edig
			  ndigs++                    // avoid initial X10
			  for c,cs,ce=r.ReadRune();ce!=io.EOF; c,cs,ce=r.ReadRune() {
				edig=c-'0'
				if edig>9 {break}
				exp=exp*10+edig
				ndigs++
				}
			  }
			// if not now on the '\0', *c must not be a digit
	
			// (this next test must be after the syntax checks)
			// if definitely more than the possible digits for format then
			// the exponent may have wrapped, so simply set it to a certain
			// over/underflow value
			if ndigs>4 {exp=6144}
			if sneg {exp=-exp}  // was negative
			} // exponent part
	
		  if (dotchar!=-1) {              // had a '.'
			digits--;                       // remove from digits count
			if (digits==0) {break}           // was dot alone: bad syntax
			exp-=clast-dotchar;      // adjust exponent
			// [the '.' can now be ignored]
			}
		  num.exponent=exp;                 // exponent is good; store it
	
		  // Here when whole string has been inspected and syntax is good
		  // cfirst->first digit or dot, clast->last digit or dot
		  err=0                          // no error possible now
	
		  // if the number of digits in the coefficient will fit in buffer
		  // then it can simply be converted to bcd8 and copied -- decFinalize
		  // will take care of leading zeros and rounding; the buffer is big
		  // enough for all canonical coefficients, including 0.00000nn...
		  ub=0
		  if digits<=61 { // [-3 allows by-4s copy]
			ic,_=r.Seek(cfirst,io.SeekStart)
			if (dotchar!=-1) {                 // a dot to worry about
				c,cs,ce=r.ReadRune()
				c1,cs,ce=r.ReadRune()
				ic,_=r.Seek(cfirst,io.SeekStart)
			if c1=='.' {                 // common canonical case
				
				buf[ub]=c-'0'           // copy leading digit
				ub++
				                            // prepare to handle rest
				}
			   else {
			     for ; ic<=clast; {          // '.' could be anywhere
				// as usual, go by fours when safe; NB it has been asserted
				// that a '.' does not have the same mask as a digit
				if ((ic<=clast-3)                             // safe for four
				 && (UBTOUI(c)&0xf0f0f0f0)==CHARMASK) {    // test four
				  UBFROMUI(ub, UBTOUI(c)&0x0f0f0f0f);      // to BCD8
				  ub+=4;
				  c+=4;
				  continue;
				  }
				if (*c=='.') {                   // found the dot
				  c++;                           // step over it ..
				  break;                         // .. and handle the rest
				  }
				*ub++=(uByte)(*c++-'0');
				}
			}
			  } // had dot
			// Now no dot; do this by fours (where safe)
			for (; c<=clast-3; c+=4, ub+=4) UBFROMUI(ub, UBTOUI(c)&0x0f0f0f0f);
			for (; c<=clast; c++, ub++) *ub=(uByte)(*c-'0');
			num.lsd=buffer+digits-1;             // record new LSD
			} // fits
	
		   else {                                // too long for buffer
			// [This is a rare and unusual case; arbitrary-length input]
			// strip leading zeros [but leave final 0 if all 0's]
			if (*cfirst=='.') cfirst++;          // step past dot at start
			if (*cfirst=='0') {                  // [cfirst always -> digit]
			  for (; cfirst<clast; cfirst++) {
				if (*cfirst!='0') {              // non-zero found
				  if (*cfirst=='.') continue;    // [ignore]
				  break;                         // done
				  }
				digits--;                        // 0 stripped
				} // cfirst
			  } // at least one leading 0
	
			// the coefficient is now as short as possible, but may still
			// be too long; copy up to Pmax+1 digits to the buffer, then
			// just record any non-zeros (set round-for-reround digit)
			for (c=cfirst; c<=clast && ub<=buffer+DECPMAX; c++) {
			  // (see commentary just above)
			  if (c<=clast-3                          // safe for four
			   && (UBTOUI(c)&0xf0f0f0f0)==CHARMASK) { // four digits
				UBFROMUI(ub, UBTOUI(c)&0x0f0f0f0f);   // to BCD8
				ub+=4;
				c+=3;                            // [will become 4]
				continue;
				}
			  if (*c=='.') continue;             // [ignore]
			  *ub++=(uByte)(*c-'0');
			  }
			ub--;                                // -> LSD
			for (; c<=clast; c++) {              // inspect remaining chars
			  if (*c!='0') {                     // sticky bit needed
				if (*c=='.') continue;           // [ignore]
				*ub=DECSTICKYTAB[*ub];           // update round-for-reround
				break;                           // no need to look at more
				}
			  }
			num.lsd=ub;                          // record LSD
			// adjust exponent for dropped digits
			num.exponent+=digits-(Int)(ub-buffer+1);
			} // too long for buffer
		  } // digits and/or dot
	
		 else {                             // no digits or dot were found
		  // only Infinities and NaNs are allowed, here
		  if (*c=='\0') break;              // nothing there is bad
		  buffer[0]=0;                      // default a coefficient of 0
		  num.lsd=buffer;                   // ..
		  if (decBiStr(c, "infinity", "INFINITY")
		   || decBiStr(c, "inf", "INF")) num.exponent=DECFLOAT_Inf;
		   else {                           // should be a NaN
			num.exponent=DECFLOAT_qNaN;     // assume quiet NaN
			if (*c=='s' || *c=='S') {       // probably an sNaN
			  num.exponent=DECFLOAT_sNaN;   // effect the 's'
			  c++;                          // and step over it
			  }
			if (*c!='N' && *c!='n') break;  // check caseless "NaN"
			c++;
			if (*c!='a' && *c!='A') break;  // ..
			c++;
			if (*c!='N' && *c!='n') break;  // ..
			c++;
			// now either nothing, or nnnn payload (no dots), expected
			// -> start of integer, and skip leading 0s [including plain 0]
			for (cfirst=c; *cfirst=='0';) cfirst++;
			if (*cfirst!='\0') {            // not empty or all-0, payload
			  // payload found; check all valid digits and copy to buffer as bcd8
			  ub=buffer;
			  for (c=cfirst;; c++, ub++) {
				if ((unsigned)(*c-'0')>9) break; // quit if not 0-9
				if (c-cfirst==DECPMAX-1) break;  // too many digits
				*ub=(uByte)(*c-'0');        // good bcd8
				}
			  if (*c!='\0') break;          // not all digits, or too many
			  num.lsd=ub-1;                 // record new LSD
			  }
			} // NaN or sNaN
		  error=0;                          // syntax is OK
		  } // digits=0 (special expected)
		break;                              // drop out
		}                                   // [for(;;) once-loop]
	}
}
